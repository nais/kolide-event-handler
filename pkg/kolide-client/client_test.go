package kolide_client_test

import (
	"os"
	"testing"

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

	t.Run("get base urls", func(t *testing.T) {
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices", kolideClient.GetApiPath("devices"))
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices/123", kolideClient.GetApiPath("devices/123"))
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices", kolideClient.GetApiPathf("devices"))
		assert.Equal(t, "https://k2.kolide.com/api/v0/devices/123", kolideClient.GetApiPathf("devices/%d", 123))
	})
}
