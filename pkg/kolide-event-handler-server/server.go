package kolide_event_handler_server

import (
	"context"
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

func (kehs *kolideEventHandlerServer) newDeviceListReceiver() (<-chan *pb.DeviceList, int) {
	kehs.channelOperationLock.Lock()
	defer kehs.channelOperationLock.Unlock()

	n := kehs.channelIDCounter
	kehs.channelIDCounter += 1

	deviceListChan := make(chan *pb.DeviceList, 50)
	kehs.deviceListReceivers[n] = deviceListChan

	return deviceListChan, n
}

func (kehs *kolideEventHandlerServer) deleteDeviceListReceiver(n int) () {
	kehs.channelOperationLock.Lock()
	defer kehs.channelOperationLock.Unlock()

	delete(kehs.deviceListReceivers, n)
}

func (kehs *kolideEventHandlerServer) broadcastDeviceList(deviceList *pb.DeviceList) {
	kehs.channelOperationLock.Lock()
	defer kehs.channelOperationLock.Unlock()

	for _, c := range kehs.deviceListReceivers {
		c <- deviceList
	}
}

func (kehs *kolideEventHandlerServer) Events(request *pb.EventsRequest, server pb.KolideEventHandler_EventsServer) error {
	deviceListReceiver, n := kehs.newDeviceListReceiver()

	for {
		select {
		case deviceList := <-deviceListReceiver:
			err := server.Send(deviceList)

			if err != nil {
				return status.Errorf(status.Code(err), "sending device list: %v", err)
			}

		case <-server.Context().Done():
			kehs.deleteDeviceListReceiver(n)
			log.Info("Events request done")
			return status.Errorf(codes.Canceled, "Events request done")

		case <-kehs.ctx.Done():
			kehs.deleteDeviceListReceiver(n)
			log.Info("kolide event handler context cancelled")
			return status.Errorf(codes.Unavailable, "server shutting down")
		}
	}
}
