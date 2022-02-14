package kolide_event_handler_server

import (
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

func New() KolideEventHandlerServer {
	return &kolideEventHandlerServer{}
}

func (kehs *kolideEventHandlerServer) Broadcast(event *pb.DeviceEvent) error {
	log.Debugf("Broadcast: %v", event)
	if kehs.client == nil {
		return nil
	}
	return kehs.client.Send(event)
}

func (kehs *kolideEventHandlerServer) Events(_ *pb.EventsRequest, server pb.KolideEventHandler_EventsServer) error {
	if kehs.client != nil {
		return status.Errorf(codes.AlreadyExists, "client already connected")
	}

	kehs.client = server

	log.Infof("Client connected to event streaming endpoint")

	<-server.Context().Done()

	log.Infof("Client disconnected from event streaming endpoint")

	kehs.client = nil

	return nil
}
