package kolide_event_handler_server

import (
	"context"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer
	deviceListChan <-chan *pb.DeviceList
	ctx            context.Context
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
}
