package kolide_event_handler_server

import (
	"github.com/nais/kolide-event-handler/pkg/pb"
)

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer

	client pb.KolideEventHandler_EventsServer
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
	Broadcast(event *pb.DeviceEvent) error
}
