package kolide_event_handler_server

import (
	"fmt"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

func New(deviceListChan <-chan *pb.DeviceList) KolideEventHandlerServer {
	return &kolideEventHandlerServer{
		deviceListChan: deviceListChan,
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
		}
	}
}
