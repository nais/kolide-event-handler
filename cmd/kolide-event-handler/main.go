package main

import (
	"errors"
	"flag"
	"github.com/nais/kolide-event-handler/pkg/pb"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	keh "github.com/nais/kolide-event-handler/pkg/kolide-event-handler"
	log "github.com/sirupsen/logrus"
)

var (
	kolideSigningSecret string
	kolideApiToken      string
)

func init() {
	flag.StringVar(&kolideSigningSecret, "kolide-signing-secret", os.Getenv("KOLIDE_SIGNING_SECRET"), "Secret for verifying webhook payloads from Kolide")
	flag.StringVar(&kolideApiToken, "kolide-api-token", os.Getenv("KOLIDE_API_TOKEN"), "API token for the Kolide API")
	flag.Parse()
}

func main() {
	httpListener, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Errorf("HTTP listener: %v", err)
		return
	}

	go startHttpServer(httpListener, kolideSigningSecret, kolideApiToken)

	grpcListener, err := net.Listen("tcp", "127.0.0.1:8081")

	if err != nil {
		log.Errorf("gRPC listener: %v", err)
		return
	}

	go startGrpcServer(grpcListener)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	log.Infof("Received %s, shutting down gracefully.", sig)
}

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
}

func startGrpcServer(listener net.Listener) {
	server := kolideEventHandlerServer{}
	grpcServer := grpc.NewServer()

	pb.RegisterKolideEventHandlerServer(grpcServer, server)

	go func() {
		log.Infof("serving gRPC on: %v", listener.Addr())
		err := grpcServer.Serve(listener)

		if err != nil {
			log.Fatalf("grcp server: %v", err)
		}
	}()
}

func startHttpServer(listener net.Listener, signingSecret, apiToken string) {
	handler := keh.New([]byte(signingSecret), apiToken)
	mux := handler.Routes()

	server := http.Server{
		Handler: mux,
	}

	go func() {
		log.Infof("serving HTTP on: %v", listener.Addr())
		err := server.Serve(listener)

		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("serving HTTP: %v", err)
		}
	}()
}

func cron() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			log.Info("Doing full Kolide device health sync")
		}
	}
}
