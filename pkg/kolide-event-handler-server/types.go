package kolide_event_handler_server

import (
	"context"
	"sync"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

type kolideEventHandlerServer struct {
	pb.UnimplementedKolideEventHandlerServer
	ctx context.Context

	deviceListFromWebhook <-chan *pb.DeviceList

	deviceListReceivers  map[int]chan *pb.DeviceList
	channelIDCounter     int
	channelOperationLock sync.Mutex
}

type KolideEventHandlerServer interface {
	pb.KolideEventHandlerServer
}
