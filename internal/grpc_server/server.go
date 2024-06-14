package grpc_server

import (
	"context"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer

	streamsLock sync.RWMutex
	streams     []chan<- *pb.DeviceEvent

	ctx context.Context
	log logrus.FieldLogger
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
	Broadcast(event *pb.DeviceEvent)
}

func New(ctx context.Context, log logrus.FieldLogger) KolideEventHandlerServer {
	return &kolideEventHandlerServer{ctx: ctx, log: log}
}

func (k *kolideEventHandlerServer) Broadcast(event *pb.DeviceEvent) {
	k.log.Debugf("Broadcast: %v", event)
	k.streamsLock.RLock()
	defer k.streamsLock.RUnlock()
	for _, stream := range k.streams {
		stream <- event
	}
}

func (k *kolideEventHandlerServer) Events(_ *pb.EventsRequest, server pb.KolideEventHandler_EventsServer) error {
	k.log.Infof("Client connected to event streaming endpoint")

	stream := make(chan *pb.DeviceEvent, 100)
	k.streamsLock.Lock()
	k.streams = append(k.streams, stream)
	k.streamsLock.Unlock()

	for {
		select {
		case <-k.ctx.Done():
			k.log.Infof("Client disconnected from event streaming endpoint")
			return nil
		case <-server.Context().Done():
			k.log.Infof("Client disconnected from event streaming endpoint")
			return nil
		case event := <-stream:
			err := server.Send(event)
			if err != nil {
				return err
			}
		}
	}
}
