package main

import (
	"context"
	"errors"
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

	keh "github.com/nais/kolide-event-handler/pkg/kolide-event-handler"
	kehs "github.com/nais/kolide-event-handler/pkg/kolide-event-handler-server"
	"github.com/nais/kolide-event-handler/pkg/pb"

	log "github.com/sirupsen/logrus"
)

var (
	kolideSigningSecret string
	kolideApiToken      string
	grpcAuthToken       string
)

func init() {
	kolideSigningSecret = os.Getenv("KOLIDE_SIGNING_SECRET")
	kolideApiToken = os.Getenv("KOLIDE_API_TOKEN")
	grpcAuthToken = os.Getenv("GRPC_AUTH_TOKEN")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	deviceListChan := make(chan *pb.DeviceList, 100)

	// HTTP Server
	httpListener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Errorf("HTTP listener: %v", err)
		return
	}

	eventHandler := keh.New(deviceListChan, []byte(kolideSigningSecret), kolideApiToken)
	go eventHandler.Cron(ctx)
	httpServer := http.Server{
		Handler: eventHandler.Routes(),
	}

	defer shutdownHttpServer(&httpServer)

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
	grpcListener, err := net.Listen("tcp", "0.0.0.0:8081")
	if err != nil {
		log.Errorf("gRPC listener: %v", err)
		return
	}

	server := kehs.New(ctx, deviceListChan)

	grpcServer := grpc.NewServer(grpc.StreamInterceptor(authenticator))
	defer func() {
		cancel()
		grpcServer.GracefulStop()
	}()

	pb.RegisterKolideEventHandlerServer(grpcServer, server)

	go func() {
		log.Infof("serving gRPC on: %v", grpcListener.Addr())
		err := grpcServer.Serve(grpcListener)

		if err != nil {
			log.Fatalf("grcp server: %v", err)
		} else {
			log.Infof("gRPC server closed")
		}
	}()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	log.Infof("Received %s, shutting down gracefully.", sig)
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
