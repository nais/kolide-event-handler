package kolide_event_handler_server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

func New(ctx context.Context, deviceListChan <-chan *pb.DeviceList) KolideEventHandlerServer {
	kehs := &kolideEventHandlerServer{
		deviceListFromWebhook: deviceListChan,
		ctx:                   ctx,
		channelIDCounter:      0,
		deviceListReceivers:   make(map[int]chan *pb.DeviceList),
	}

	go kehs.WatchDeviceListChannel(ctx)

	return kehs
}

func (kehs *kolideEventHandlerServer) newDeviceListReceiver() (<-chan *pb.DeviceList, int) {
	kehs.mapLock.Lock()
	defer kehs.mapLock.Unlock()

	n := kehs.channelIDCounter
	kehs.channelIDCounter += 1

	deviceListChan := make(chan *pb.DeviceList, 50)
	kehs.deviceListReceivers[n] = deviceListChan

	return deviceListChan, n
}

func (kehs *kolideEventHandlerServer) deleteDeviceListReceiver(n int) () {
	kehs.mapLock.Lock()
	defer kehs.mapLock.Unlock()

	delete(kehs.deviceListReceivers, n)
}

func (kehs *kolideEventHandlerServer) broadcastDeviceList(deviceList *pb.DeviceList) {
	kehs.mapLock.Lock()
	defer kehs.mapLock.Unlock()

	for _, c := range kehs.deviceListReceivers {
		c <- deviceList
	}
}

func (kehs *kolideEventHandlerServer) WatchDeviceListChannel(ctx context.Context) {
	for {
		select {
		case deviceList := <-kehs.deviceListFromWebhook:
			log.Infof("broadcasting deviceList to receivers")
			kehs.broadcastDeviceList(deviceList)
		case <-ctx.Done():
			log.Infof("stopping watchDeviceListChannel")
			return
		}
	}
}

func (kehs *kolideEventHandlerServer) Events(request *pb.EventsRequest, server pb.KolideEventHandler_EventsServer) error {
	deviceListReceiver, n := kehs.newDeviceListReceiver()

	for {
		select {
		case deviceList := <-deviceListReceiver:
			log.Infof("sending device list to %d", n)
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
