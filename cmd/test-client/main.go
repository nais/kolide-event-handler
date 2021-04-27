package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

var (
	server string
)

type ClientInterceptor struct {
	RequireTLS bool
	Token      string
}

func (c *ClientInterceptor) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": c.Token,
	}, nil
}

func (c *ClientInterceptor) RequireTransportSecurity() bool {
	return c.RequireTLS
}

func init() {
	flag.StringVar(&server, "server", "127.0.0.1:8081", "target grpc server")
	flag.Parse()
}

func main() {
	interceptor := &ClientInterceptor{
		RequireTLS: false,
		Token:      os.Getenv("GRPC_AUTH_TOKEN"),
	}

	cred := credentials.NewTLS(&tls.Config{})
	conn, err := grpc.Dial(server, grpc.WithTransportCredentials(cred), grpc.WithPerRPCCredentials(interceptor))
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
	for {
		events, err := s.Events(ctx, &pb.EventsRequest{})
		if err != nil {
			log.Errorf("calling rpc: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		log.Infof("connected to %v", conn.Target())

		for {
			event, err := events.Recv()
			if err != nil {
				if status.Code(err) == codes.Unavailable {
					log.Warnf("err: %+v", err)
					time.Sleep(1 * time.Second)
					break
				} else {
					log.Errorf("receiving event: %v", err)
					return
				}
			}

			eventJson, err := json.Marshal(event)
			if err != nil {
				log.Errorf("marshalling event: %v", err)
			}

			log.Infof("event received: %s", eventJson)
		}
	}
}
