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
		{[]string{}, keh.SEVERITY_WARNING, keh.DURATION_WARNING},
		{[]string{"foo", "bar"}, keh.SEVERITY_WARNING, keh.DURATION_WARNING},
		{[]string{"foo", "notice"}, keh.SEVERITY_NOTICE, keh.DURATION_NOTICE},
		{[]string{"warning", "notice", "danger"}, keh.SEVERITY_DANGER, keh.DURATION_DANGER},
		{[]string{"notice"}, keh.SEVERITY_NOTICE, keh.DURATION_NOTICE},
		{[]string{"warning"}, keh.SEVERITY_WARNING, keh.DURATION_WARNING},
		{[]string{"danger"}, keh.SEVERITY_DANGER, keh.DURATION_DANGER},
		{[]string{"critical"}, keh.SEVERITY_CRITICAL, keh.DURATION_CRITICAL},
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
