package kolide_event_handler

import (
	"time"

	kolide_client "github.com/nais/kolide-event-handler/pkg/kolide-client"
)

type KolideEventHandler struct {
	signingSecret []byte
	apiClient     *kolide_client.KolideClient
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

type Severity int

const (
	SEVERITY_INFO Severity = iota
	SEVERITY_NOTICE
	SEVERITY_WARNING
	SEVERITY_DANGER
	SEVERITY_CRITICAL
)

type SeverityDuration time.Duration

const (
	DURATION_NOTICE   = time.Hour * 24 * 7
	DURATION_WARNING  = time.Hour * 24 * 2
	DURATION_DANGER   = time.Hour
	DURATION_CRITICAL = 0
	DURATION_UNKNOWN  = time.Hour * 24 * 30
)
