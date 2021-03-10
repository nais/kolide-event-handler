package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

var (
	server      string
	TLSKeyPath  string
	TLSCertPath string
	TLSCAPath   string
	TLSExtraPath   string
)

func init() {
	flag.StringVar(&server, "server", "127.0.0.1:8081", "target grpc server")
	flag.StringVar(&TLSCertPath, "tls-cert", "/var/run/secrets/tls/tls.crt", "Server certificate path")
	flag.StringVar(&TLSKeyPath, "tls-key", "/var/run/secrets/tls/tls.key", "Server key path")
	flag.StringVar(&TLSCAPath, "tls-ca", "/var/run/secrets/ca/ca.crt", "Client CA path")
	flag.StringVar(&TLSExtraPath, "tls-extra", "/var/run/secrets/ca/extra.crt", "Client extra path")
	flag.Parse()
}

func main() {
	creds, err := loadTLS(TLSCertPath, TLSKeyPath, TLSCAPath, TLSExtraPath)
	if err != nil {
		log.Errorf("loading tls: %v", err)
		return
	}

	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(creds))
	//conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("connecting to grpc server: %v", err)
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Errorf("closing grpc connection: %v", err)
		}
	}()

	s := pb.NewKolideEventHandlerClient(conn)

	ctx := context.Background()
	events, err := s.Events(ctx, &pb.EventsRequest{})
	if err != nil {
		log.Errorf("calling rpc: %v", err)
		return
	}

	for {
		event, err := events.Recv()
		if err != nil {
			log.Errorf("receiving event: %v", err)
			return
		}

		log.Infof("event received: %+v", event)
	}
}

func loadTLS(certPath, keyPath, caPath, extraPath string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("loading keypair: %w", err)
	}

	ca, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("readin ca file: %w", err)
	}

	caPool := x509.NewCertPool()

	if !caPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("unable to add ca to ca pool")
	}

	extra, err := ioutil.ReadFile(extraPath)
	if err != nil {
		return nil, fmt.Errorf("readin ca file: %w", err)
	}
	if !caPool.AppendCertsFromPEM(extra) {
		return nil, fmt.Errorf("unable to add extra to ca pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      caPool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
