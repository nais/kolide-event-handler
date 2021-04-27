package kolide_event_handler_test

import (
	"strings"
	"testing"
	"time"

	kolide_client "github.com/nais/kolide-event-handler/pkg/kolide-client"
	keh "github.com/nais/kolide-event-handler/pkg/kolide-event-handler"
	"github.com/stretchr/testify/assert"
)

func TestEventHandler(t *testing.T) {
	var tagTests = []struct {
		tags     []string
		severity keh.Severity
		duration time.Duration
	}{
		{[]string{}, keh.SeverityWarning, keh.DurationWarning},
		{[]string{"foo", "bar"}, keh.SeverityWarning, keh.DurationWarning},
		{[]string{"foo", "notice"}, keh.SeverityNotice, keh.DurationNotice},
		{[]string{"warning", "notice", "danger"}, keh.SeverityDanger, keh.DurationDanger},
		{[]string{"notice"}, keh.SeverityNotice, keh.DurationNotice},
		{[]string{"warning"}, keh.SeverityWarning, keh.DurationWarning},
		{[]string{"danger"}, keh.SeverityDanger, keh.DurationDanger},
		{[]string{"critical"}, keh.SeverityCritical, keh.DurationCritical},
	}

	for _, tt := range tagTests {
		t.Run(strings.Join(tt.tags, ", "), func(t *testing.T) {
			check := kolide_client.Check{
				Tags: tt.tags,
			}

			severity := keh.GetSeverity(check)

			assert.Equal(t, tt.severity, severity)
			assert.Equal(t, tt.duration, keh.GetGraceTime(severity))
		})
	}
}
