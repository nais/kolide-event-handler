package kolide_event_handler_server

import (
	"context"
	"sync"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer
	ctx context.Context

	deviceEventChan <-chan *pb.DeviceEvent

	deviceEventReceivers map[int]chan *pb.DeviceEvent
	channelIDCounter     int
	mapLock              sync.Mutex
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
}
