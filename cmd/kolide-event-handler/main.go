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

	"github.com/nais/kolide-event-handler/pkg/kolide"
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

const (
	kolideGetDeviceTimeout  = 10 * time.Second
	kolideGetDevicesTimeout = 4 * time.Minute
	kolideFullSyncInterval  = 5 * time.Minute
)

func init() {
	kolideSigningSecret = os.Getenv("KOLIDE_SIGNING_SECRET")
	kolideApiToken = os.Getenv("KOLIDE_API_TOKEN")
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

	failures := make(chan keh.KolideEventFailure, 1000)

	// HTTP Server
	httpListener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		return fmt.Errorf("HTTP listener: %v", err)
	}

	client := kolide.New(kolideApiToken)

	eventHandler := keh.New(client, failures, []byte(kolideSigningSecret))
	httpServer := &http.Server{
		Handler: eventHandler.Routes(),
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
	grpcListener, err := net.Listen("tcp", "0.0.0.0:8081")
	if err != nil {
		return fmt.Errorf("gRPC listener: %v", err)
	}

	server := kehs.New()

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

	process := func(ev keh.KolideEventFailure) error {
		ctx, cancel := context.WithTimeout(ctx, kolideGetDeviceTimeout)
		defer cancel()
		device, err := client.GetDevice(ctx, uint64(ev.Data.DeviceId))
		if err != nil {
			return err
		}
		return server.Broadcast(device.Event())
	}

	fullSync := func() error {
		ctx, cancel := context.WithTimeout(ctx, kolideGetDevicesTimeout)
		defer cancel()
		devices, err := client.GetDevices(ctx)
		if err != nil {
			return fmt.Errorf("get devices: %w", err)
		}
		for _, device := range devices {
			err = server.Broadcast(device.Event())
			if err != nil {
				return fmt.Errorf("send device event: %s", err)
			}
		}
		return nil
	}

	fullSyncTimer := time.NewTimer(10 * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			grpcServer.Stop()
			httpServer.Close()
			return nil

		case ev := <-failures:
			err := process(ev)
			if err != nil {
				log.Errorf("process event: %s", err)
			}

		case <-fullSyncTimer.C:
			then := time.Now()
			log.Debugf("Synchronizing against Kolide...")
			err := fullSync()
			if err != nil {
				log.Errorf("full sync: %s", err)
			}
			log.Debugf("Finished synchronizing against Kolide in %s", time.Since(then))
			fullSyncTimer.Reset(kolideFullSyncInterval)
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
