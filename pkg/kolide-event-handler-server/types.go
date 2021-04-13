package kolide_event_handler_server

import (
	"context"
	"sync"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer
	deviceListChan <-chan *pb.DeviceList
	ctx            context.Context

	deviceListReceivers  map[int]chan<- *pb.DeviceList
	channelIDCounter     int
	channelOperationLock sync.Mutex
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
}
