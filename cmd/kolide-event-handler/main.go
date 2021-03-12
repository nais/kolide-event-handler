package main

import (
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

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
	deviceListChan := make(chan *pb.DeviceList, 100)
	httpListener, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		log.Errorf("HTTP listener: %v", err)
		return
	}

	handler := keh.New(deviceListChan, []byte(kolideSigningSecret), kolideApiToken)

	go startHttpServer(httpListener, handler.Routes())

	grpcListener, err := net.Listen("tcp", "127.0.0.1:8081")

	if err != nil {
		log.Errorf("gRPC listener: %v", err)
		return
	}

	server := kehs.New(deviceListChan)
	go startGrpcServer(grpcListener, server)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	sig := <-interrupt
	log.Infof("Received %s, shutting down gracefully.", sig)
}

func startGrpcServer(listener net.Listener, server kehs.KolideEventHandlerServer) {
	grpcServer := grpc.NewServer(grpc.StreamInterceptor(authenticator))

	pb.RegisterKolideEventHandlerServer(grpcServer, server)

	go func() {
		log.Infof("serving gRPC on: %v", listener.Addr())
		err := grpcServer.Serve(listener)

		if err != nil {
			log.Fatalf("grcp server: %v", err)
		}
	}()
}

func authenticator(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	md, _ := metadata.FromIncomingContext(ss.Context())

	if strings.Join(md.Get("authorization"), "") != grpcAuthToken {
		return status.Errorf(codes.Unauthenticated, "incorrect authorization")
	}

	log.Infof(info.FullMethod)
	log.Infof("%+v", md)
	return handler(srv, ss)
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
