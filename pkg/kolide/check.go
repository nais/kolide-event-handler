package kolide

import (
	"fmt"
	"strings"
	"time"

	"github.com/nais/kolide-event-handler/pkg/pb"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (check Check) Severity() Severity {
	var severity, max Severity = -1, -1

	for _, tag := range check.Tags {
		switch strings.ToLower(tag) {
		case "info":
			severity = SeverityInfo
		case "notice":
			severity = SeverityNotice
		case "warning":
			severity = SeverityWarning
		case "danger":
			severity = SeverityDanger
		case "critical":
			severity = SeverityCritical
		}

		if severity > max {
			max = severity
		}
	}

	if max == -1 {
		log.Warnf("Check missing a severity tag: %+v", check)
		max = SeverityWarning
	}

	return max
}

func (severity Severity) GraceTime() time.Duration {
	switch severity {
	case SeverityNotice:
		return DurationNotice
	case SeverityWarning:
		return DurationWarning
	case SeverityDanger:
		return DurationDanger
	case SeverityCritical:
		return DurationCritical
	default:
		return DurationUnknown
	}
}

func (failure *DeviceFailure) Health() pb.Health {
	if failure == nil || failure.Ignored || failure.ResolvedAt != nil {
		return pb.Health_Healthy
	}

	if failure.Check == nil {
		log.Errorf("BUG: malformed failure from Kolide API: failure=%d; checkID=%d", failure.Id, failure.CheckId)
		return pb.Health_Healthy
	}

	// Ignore INFO checks
	severity := failure.Check.Severity()
	if severity == SeverityInfo {
		return pb.Health_Healthy
	}

	graceTime := severity.GraceTime()
	if graceTime == DurationUnknown {
		log.Errorf("DurationUnknown grace time for check %d, with tags: %+v", failure.CheckId, failure.Check.Tags)
	}

	// Deny by default if check time is unknown; might have been a long time ago
	if failure.Timestamp == nil {
		return pb.Health_Unhealthy
	}

	deadline := failure.Timestamp.Add(graceTime)
	if time.Now().After(deadline) {
		return pb.Health_Unhealthy
	}

	return pb.Health_Healthy
}

const MaxTimeSinceKolideLastSeen = 240 * time.Hour

// If one check fails, the device is unhealthy.
func (device *Device) Health() (pb.Health, string) {

	// Allow only registered devices
	if len(device.AssignedOwner.Email) == 0 {
		return pb.Health_Unhealthy, "Kolide does not know who owns this device"
	}

	// Devices must phone home regularly
	lastSeen := time.Time{}
	if device.LastSeenAt != nil {
		lastSeen = *device.LastSeenAt
	}
	deadline := lastSeen.Add(MaxTimeSinceKolideLastSeen)
	if time.Now().After(deadline) {
		msg := fmt.Sprintf("Kolide's information about this device is out of date")
		return pb.Health_Unhealthy, msg
	}

	// Any failure means device failure
	for _, failure := range device.Failures {
		if failure.Health() == pb.Health_Unhealthy {
			return pb.Health_Unhealthy, failure.Title
		}
	}

	return pb.Health_Healthy, ""
}

func (device *Device) Event() *pb.DeviceEvent {
	health, msg := device.Health()
	platform := device.Platform
	if len(platform) == 0 {
		platform = device.OSDetails.Platform
	}
	return &pb.DeviceEvent{
		Timestamp: timestamppb.Now(),
		Serial:    device.Serial,
		Platform:  device.Platform,
		State:     health,
		Message:   msg,
	}
}
