package kolide_event_handler_server

import "github.com/nais/kolide-event-handler/pkg/pb"

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer
	deviceListChan <-chan *pb.DeviceList
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
}
