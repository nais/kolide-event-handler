package kolide_client_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	kolide_client "github.com/nais/kolide-event-handler/pkg/kolide-client"
	"github.com/stretchr/testify/assert"
)

func TestClient(t *testing.T) {
	kolideClient := kolide_client.New(os.Getenv("KOLIDE_API_TOKEN"))

	t.Run("get device", func(t *testing.T) {
		t.Skip()
		_, err := kolideClient.GetDevice(25215)
		assert.NoError(t, err)
	})

	t.Run("get check", func(t *testing.T) {
		t.Skip()
		_, err := kolideClient.GetCheck(27680)
		assert.NoError(t, err)
	})

	t.Run("get failure", func(t *testing.T) {
		t.Skip()
		deviceFailure, err := kolideClient.GetDeviceFailure(27066, 123)
		t.Logf("device: %+v", deviceFailure)
		assert.Error(t, err)
	})

	t.Run("rate limit test", func(t *testing.T) {
		tests := []struct {
			name       string
			header     http.Header
			retryAfter time.Duration
			compareFunc func(got, want time.Duration) bool
		}{
			{
				name:   "no headers should give 0",
				header: http.Header{ },
				retryAfter: 0,
			},
			{
				name: "correct header should give value",
				header: http.Header{
					"Retry-After":[]string{"5"},
				},
				retryAfter: 5*time.Second,
			},
			{
				name: "invalid header should give default value",
				header: http.Header{
					"Retry-After":[]string{"a"},
				},
				retryAfter: kolide_client.DefaultRetryAfter,
			},
			{
				name: "negative header should give default value",
				header: http.Header{
					"Retry-After":[]string{"-4"},
				},
				retryAfter: kolide_client.DefaultRetryAfter,
			},
			{
				name: "retry-after in the past should give default",
				header: http.Header{
					"Retry-After":[]string{time.Now().Add(-time.Hour).Format(time.RFC1123)},
				},
				retryAfter: kolide_client.DefaultRetryAfter,
			},
			{
				name: "retry-after in the future should give delta",
				header: http.Header{
					"Retry-After":[]string{time.Now().Add(time.Hour).Format(time.RFC1123)},
				},
				retryAfter: time.Hour,
				compareFunc: func(got, want time.Duration) bool {
					return want - got <= time.Second
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := kolide_client.GetRetryAfter(tt.header)
				if tt.compareFunc != nil {
					if !tt.compareFunc(got, tt.retryAfter) {
						t.Errorf("GetRetryAfter() = %v, want %v (using tt.compareFunc)", got, tt.retryAfter)
					}
				} else {
					if got != tt.retryAfter {
						t.Errorf("GetRetryAfter() = %v, want %v", got, tt.retryAfter)
					}
				}
			})
		}
	})

	t.Run("get devices", func(t *testing.T) {
		t.Skip()
		devices, err := kolideClient.GetDevices()
		assert.NoError(t, err)
		t.Logf("devices: %+v", len(devices))
		t.Logf("device sample: %+v", devices[len(devices)-1])
	})

	t.Run("get base urls", func(t *testing.T) {
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices", kolideClient.GetApiPath("devices"))
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices/123", kolideClient.GetApiPath("devices/123"))
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices", kolideClient.GetApiPathf("devices"))
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices/123", kolideClient.GetApiPathf("devices/%d", 123))
	})
}
