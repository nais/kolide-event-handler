package main

import (
	"context"
	"flag"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

var (
	server      string
)

func init() {
	flag.StringVar(&server, "server", "127.0.0.1:8081", "target grpc server")
	flag.Parse()
}

func main() {
	conn, err := grpc.Dial(server, grpc.WithInsecure())
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