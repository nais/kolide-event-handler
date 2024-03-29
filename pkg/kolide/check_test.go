package kolide_test

import (
	"strings"
	"testing"
	"time"

	kolideclient "github.com/nais/kolide-event-handler/pkg/kolide"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	tagTests := []struct {
		tags     []string
		severity kolideclient.Severity
		duration time.Duration
	}{
		{[]string{}, kolideclient.SeverityWarning, kolideclient.DurationWarning},
		{[]string{"foo", "bar"}, kolideclient.SeverityWarning, kolideclient.DurationWarning},
		{[]string{"foo", "notice"}, kolideclient.SeverityNotice, kolideclient.DurationNotice},
		{[]string{"warning", "notice", "danger"}, kolideclient.SeverityDanger, kolideclient.DurationDanger},
		{[]string{"notice"}, kolideclient.SeverityNotice, kolideclient.DurationNotice},
		{[]string{"warning"}, kolideclient.SeverityWarning, kolideclient.DurationWarning},
		{[]string{"danger"}, kolideclient.SeverityDanger, kolideclient.DurationDanger},
		{[]string{"critical"}, kolideclient.SeverityCritical, kolideclient.DurationCritical},
	}

	for _, tt := range tagTests {
		t.Run(strings.Join(tt.tags, ", "), func(t *testing.T) {
			check := kolideclient.Check{
				Tags: tt.tags,
			}

			severity := check.Severity()

			assert.Equal(t, tt.severity, severity)
			assert.Equal(t, tt.duration, severity.GraceTime())
		})
	}
}
