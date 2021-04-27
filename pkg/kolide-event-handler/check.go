package kolide_event_handler

import (
	"github.com/nais/kolide-event-handler/pkg/pb"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

func GetSeverity(check *pb.Check) Severity {
	var severity, mostSevereSeverity Severity = -1, -1

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

		if severity > mostSevereSeverity {
			mostSevereSeverity = severity
		}
	}

	if mostSevereSeverity == -1 {
		log.Warnf("Check missing a severity tag: %+v", check)
		mostSevereSeverity = SeverityWarning
	}

	return mostSevereSeverity
}

func GetGraceTime(severity Severity) time.Duration {
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
		log.Errorf("Unknown severity: %v", severity)
		return DurationUnknown
	}
}
