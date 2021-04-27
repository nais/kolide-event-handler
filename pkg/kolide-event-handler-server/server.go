package kolide_event_handler_server

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

func New(ctx context.Context, deviceEventChan <-chan *pb.DeviceEvent) KolideEventHandlerServer {
	kehs := &kolideEventHandlerServer{
		ctx:                  ctx,
		deviceEventChan:      deviceEventChan,
		channelIDCounter:     0,
		deviceEventReceivers: make(map[int]chan *pb.DeviceEvent),
	}

	go kehs.WatchDeviceEventChannel(ctx)

	return kehs
}

func (kehs *kolideEventHandlerServer) newDeviceEventReceiver() (<-chan *pb.DeviceEvent, int) {
	kehs.mapLock.Lock()
	defer kehs.mapLock.Unlock()

	n := kehs.channelIDCounter
	kehs.channelIDCounter += 1

	deviceEventChan := make(chan *pb.DeviceEvent, 50)
	kehs.deviceEventReceivers[n] = deviceEventChan

	return deviceEventChan, n
}

func (kehs *kolideEventHandlerServer) deleteDeviceEventReceiver(n int) {
	kehs.mapLock.Lock()
	defer kehs.mapLock.Unlock()

	delete(kehs.deviceEventReceivers, n)
}

func (kehs *kolideEventHandlerServer) broadcastDeviceEvent(deviceEvent *pb.DeviceEvent) {
	kehs.mapLock.Lock()
	defer kehs.mapLock.Unlock()

	for n, c := range kehs.deviceEventReceivers {
		log.Debugf("send deviceEvent to receiver %d", n)
		c <- deviceEvent
	}
}

func (kehs *kolideEventHandlerServer) WatchDeviceEventChannel(ctx context.Context) {
	for {
		select {
		case deviceEvent := <-kehs.deviceEventChan:
			log.Debugf("broadcast deviceEvent to receivers")
			kehs.broadcastDeviceEvent(deviceEvent)
		case <-ctx.Done():
			log.Infof("stop watchDeviceEventChannel")
			return
		}
	}
}

func (kehs *kolideEventHandlerServer) Events(_ *pb.EventsRequest, server pb.KolideEventHandler_EventsServer) error {
	deviceEventReceiver, n := kehs.newDeviceEventReceiver()

	for {
		select {
		case deviceEvent := <-deviceEventReceiver:
			log.Debugf("send deviceEvent to %d", n)
			err := server.Send(deviceEvent)

			if err != nil {
				return status.Errorf(status.Code(err), "send deviceEvent: %v", err)
			}

		case <-server.Context().Done():
			kehs.deleteDeviceEventReceiver(n)
			log.Info("Events request done")
			return status.Errorf(codes.Canceled, "Events request done")

		case <-kehs.ctx.Done():
			kehs.deleteDeviceEventReceiver(n)
			log.Info("kolide event handler context cancelled")
			return status.Errorf(codes.Unavailable, "server shutting down")
		}
	}
}
