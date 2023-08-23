package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	kolideclient "github.com/nais/kolide-event-handler/pkg/kolide"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

var server string

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
	apiToken := os.Getenv("KOLIDE_API_TOKEN")
	if len(apiToken) == 0 {
		log.Errorf("env KOLIDE_API_TOKEN not found, aborting")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Graceful CTRL-C handling
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func(cancel func()) {
		sig := <-signals
		log.Infof("Received signal %s, exiting...", sig)
		cancel()
	}(cancel)

	go Cron(ctx, apiToken)

	interceptor := &ClientInterceptor{
		RequireTLS: false,
		Token:      os.Getenv("GRPC_AUTH_TOKEN"),
	}

	cred := credentials.NewTLS(&tls.Config{})
	conn, err := grpc.DialContext(ctx, server, grpc.WithTransportCredentials(cred), grpc.WithPerRPCCredentials(interceptor))
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

eventloop:
	for {
		events, err := s.Events(ctx, &pb.EventsRequest{})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Infof("program finished")
				break
			}

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
					break eventloop
				}
			}

			log.Infof("event received: %+v", event)
		}
	}

	log.Info("bye")
}

const (
	FullSyncInterval = 5 * time.Minute
	FullSyncTimeout  = 3 * time.Minute // Must not be greater than FullSyncInterval
)

func Cron(programContext context.Context, apiToken string) {
	ticker := time.NewTicker(time.Second * 1)
	apiClient := kolideclient.New(apiToken)

	for {
		select {
		case <-ticker.C:
			ticker.Reset(FullSyncInterval)
			log.Info("Doing full Kolide device health sync")
			ctx, cancel := context.WithTimeout(programContext, FullSyncTimeout)
			devices, err := apiClient.GetDevices(ctx)
			cancel()
			if err != nil {
				log.Errorf("getting devies: %v", err)
			}

			devicesJson, err := json.Marshal(devices)
			if err != nil {
				log.Errorf("marshal json: %v", err)
			}

			log.Infof("%s", devicesJson)

		case <-programContext.Done():
			log.Infof("stopping cron")
			return
		}
	}
}
