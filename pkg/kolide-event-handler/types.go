package kolide_event_handler

import (
	"time"

	kolideclient "github.com/nais/kolide-event-handler/pkg/kolide-client"
	"github.com/nais/kolide-event-handler/pkg/pb"
)

type KolideEventHandler struct {
	signingSecret []byte
	apiClient     *kolideclient.KolideClient
	listChan      chan<- *pb.DeviceList
}

type KolideEventFailureData struct {
	CheckId   uint64 `json:"check_id"`
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

type Severity int

const (
	SeverityInfo Severity = iota
	SeverityNotice
	SeverityWarning
	SeverityDanger
	SeverityCritical
)

type SeverityDuration time.Duration

const (
	DurationNotice   = time.Hour * 24 * 7
	DurationWarning  = time.Hour * 24 * 2
	DurationDanger   = time.Hour
	DurationCritical = 0
	DurationUnknown  = time.Hour * 24 * 30
)
