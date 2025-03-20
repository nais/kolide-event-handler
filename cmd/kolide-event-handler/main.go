package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/nais/kolide-event-handler/internal/grpc_server"
	"github.com/nais/kolide-event-handler/internal/webhook_handler"
	"github.com/nais/kolide-event-handler/pkg/pb"

	log "github.com/sirupsen/logrus"
)

var (
	kolideSigningSecret string
	grpcAuthToken       string
)

func init() {
	kolideSigningSecret = os.Getenv("KOLIDE_SIGNING_SECRET")
	grpcAuthToken = os.Getenv("GRPC_AUTH_TOKEN")
}

func main() {
	err := run()
	if err != nil {
		log.Errorf("fatal: %s", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	failures := make(chan webhook_handler.KolideEventFailure, 1000)

	// HTTP Server
	httpListener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		return fmt.Errorf("HTTP listener: %v", err)
	}

	eventHandler := webhook_handler.New(failures, []byte(kolideSigningSecret), log.WithField("component", "webhook_handler"))
	httpServer := &http.Server{
		Handler:           eventHandler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	defer shutdownHttpServer(httpServer)

	go func() {
		log.Infof("serving HTTP on: %v", httpListener.Addr())
		err := httpServer.Serve(httpListener)

		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("serving HTTP: %v", err)
		} else {
			log.Infof("HTTP server closed")
		}
	}()

	// GRPC Server
	grpcListener, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		return fmt.Errorf("gRPC listener: %v", err)
	}

	server := grpc_server.New(ctx, log.WithField("component", "grpc_server"))

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(authenticator))
	defer func() {
		cancel()
		grpcServer.GracefulStop()
	}()

	pb.RegisterKolideEventHandlerServer(grpcServer, server)

	go func() {
		log.Infof("gRPC server starting on %v", grpcListener.Addr())
		err := grpcServer.Serve(grpcListener)
		cancel()

		if err != nil {
			log.Fatalf("gRPC server: %v", err)
		} else {
			log.Infof("gRPC server closed")
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signals
		log.Infof("Received %s, shutting down.", sig)
		cancel()
	}()

	process := func(ev webhook_handler.KolideEventFailure) {
		event := &pb.DeviceEvent{
			Timestamp:  timestamppb.Now(),
			ExternalID: fmt.Sprint(ev.Id),
		}
		server.Broadcast(event)
	}

	for {
		select {
		case <-ctx.Done():
			grpcServer.Stop()
			if err := httpServer.Close(); err != nil {
				log.WithError(err).Errorf("closing HTTP server")
			}
			return nil

		case ev := <-failures:
			process(ev)
		}
	}
}

func authenticator(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, _ := metadata.FromIncomingContext(ss.Context())

	if strings.Join(md.Get("authorization"), "") != grpcAuthToken {
		return status.Errorf(codes.Unauthenticated, "incorrect authorization")
	}

	return handler(srv, ss)
}

func shutdownHttpServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Errorf("shutting down http server: %v", err)
	}

	if ctx.Err() != nil {
		log.Errorf("shutdown context error: %v", err)
	}
}
