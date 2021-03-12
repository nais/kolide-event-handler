package kolide_event_handler_server

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

func New(ctx context.Context, deviceListChan <-chan *pb.DeviceList) KolideEventHandlerServer {
	return &kolideEventHandlerServer{
		deviceListChan: deviceListChan,
		ctx:            ctx,
	}
}

func (kehs *kolideEventHandlerServer) Events(request *pb.EventsRequest, server pb.KolideEventHandler_EventsServer) error {
	for {
		select {
		case deviceList := <-kehs.deviceListChan:
			err := server.Send(deviceList)

			if err != nil {
				return fmt.Errorf("sending device list: %w", err)
			}
		case <-kehs.ctx.Done():
			log.Info("kolide event handler context cancelled")
			return status.Errorf(codes.Unavailable, "server shutting down")
		}
	}
}
