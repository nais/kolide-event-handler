package kolide_event_handler

import (
	"time"

	"github.com/nais/kolide-event-handler/pkg/pb"
)

type KolideEventHandler struct {
	signingSecret     []byte
	notifyDeviceEvent chan<- *pb.DeviceEvent
}

type KolideEventFailureData struct {
	CheckId   int    `json:"check_id"`
	FailureId int    `json:"failure_id"`
	Title     string `json:"title"`
	DeviceId  int    `json:"device_id"`
}

type KolideEventFailure struct {
	Id        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      KolideEventFailureData `json:"data"`
	KolideEvent
}

type KolideEvent struct {
	Event string `json:"event"`
}
