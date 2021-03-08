package main

import (
	"errors"
	"flag"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	keh "github.com/nais/kolide-event-handler/pkg/kolide-event-handler"
	kehs "github.com/nais/kolide-event-handler/pkg/kolide-event-handler-server"
	"github.com/nais/kolide-event-handler/pkg/pb"

	log "github.com/sirupsen/logrus"
)

var (
	kolideSigningSecret string
	kolideApiToken      string
	tlsKeyPath          string
	tlsCertPath         string
	tlsCAPath           string
)

func init() {
	flag.StringVar(&kolideSigningSecret, "kolide-signing-secret", os.Getenv("KOLIDE_SIGNING_SECRET"), "Secret for verifying webhook payloads from Kolide")
	flag.StringVar(&kolideApiToken, "kolide-api-token", os.Getenv("KOLIDE_API_TOKEN"), "API token for the Kolide API")
	flag.StringVar(&tlsCertPath, "tls-cert", "/var/run/secrets/tls/tls.crt", "Server certificate path")
	flag.StringVar(&tlsKeyPath, "tls-key", "/var/run/secrets/tls/tls.key", "Server key path")
	flag.StringVar(&tlsCAPath, "tls-ca", "/var/run/secrets/ca/ca.crt", "Client CA path")
	flag.Parse()
}

func main() {
	deviceListChan := make(chan *pb.DeviceList, 100)
	// some test data, TODO remove later
	for i := 0; i < 10; i++ {
		deviceListChan<-&pb.DeviceList{
			Devices: []*pb.DeviceHealthEvent{
				{
					DeviceId: uint64(i),
					Health:   false,
					LastSeen: nil,
					Serial:   "serial",
					Username: "username",
				},
			},
		}
	}

	httpListener, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Errorf("HTTP listener: %v", err)
		return
	}

	handler := keh.New(deviceListChan, []byte(kolideSigningSecret), kolideApiToken)

	go startHttpServer(httpListener, handler.Routes())

	creds, err := loadTLS(tlsCertPath, tlsKeyPath, tlsCAPath)
	if err != nil {
		log.Errorf("reading credentials from file: %v", err)
		return
	}

	grpcListener, err := net.Listen("tcp", "127.0.0.1:8081")

	if err != nil {
		log.Errorf("gRPC listener: %v", err)
		return
	}

	server := kehs.New(deviceListChan)
	go startGrpcServer(grpcListener, server, creds)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	log.Infof("Received %s, shutting down gracefully.", sig)
}

func startGrpcServer(listener net.Listener, server kehs.KolideEventHandlerServer, creds credentials.TransportCredentials) {
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(LogClientCertSubj),
		grpc.StreamInterceptor(StreamLogClientCertSubj),
	)

	pb.RegisterKolideEventHandlerServer(grpcServer, server)

	go func() {
		log.Infof("serving gRPC on: %v", listener.Addr())
		err := grpcServer.Serve(listener)

		if err != nil {
			log.Fatalf("grcp server: %v", err)
		}
	}()
}

func startHttpServer(listener net.Listener, handler http.Handler) {
	server := http.Server{
		Handler: handler,
	}

	go func() {
		log.Infof("serving HTTP on: %v", listener.Addr())
		err := server.Serve(listener)

		if !errors.Is(err, http.ErrServerClosed) {
			log.Errorf("serving HTTP: %v", err)
		}
	}()
}

/*
func cron() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			log.Info("Doing full Kolide device health sync")
		}
	}
}
*/