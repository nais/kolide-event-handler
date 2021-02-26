package kolide_client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type KolideClient struct {
	baseUrl string
	client  *http.Client
}

func New(token string) *KolideClient {

	t := Transport{
		Token: token,
	}

	return &KolideClient{
		baseUrl: "https://k2.kolide.com/api/v0",
		client:  t.Client(),
	}
}

func (kc *KolideClient) Get(path string) (*http.Response, error) {
	return kc.client.Get(kc.GetApiPath(path))
}

func (kc *KolideClient) GetApiPath(path string) string {
	return fmt.Sprintf("%s/%s", kc.baseUrl, path)
}

func (kc *KolideClient) GetApiPathf(path string, args ...interface{}) string {
	return kc.GetApiPath(fmt.Sprintf(path, args...))
}

func (kc *KolideClient) GetDevice(deviceId int) (*Device, error) {
	response, err := kc.Get(fmt.Sprintf("devices/%d", deviceId))

	if err != nil {
		return nil, fmt.Errorf("getting client: %w", err)
	}

	var device Device

	err = json.NewDecoder(response.Body).Decode(&device)

	if err != nil {
		return nil, fmt.Errorf("decoding device: %w", err)
	}

	return &device, nil
}

func (kc *KolideClient) GetCheck(checkId int) (*Check, error) {
	response, err := kc.Get(fmt.Sprintf("checks/%d", checkId))

	if err != nil {
		return nil, fmt.Errorf("getting check: %w", err)
	}

	var check Check

	err = json.NewDecoder(response.Body).Decode(&check)

	if err != nil {
		return nil, fmt.Errorf("decoding check: %w", err)
	}

	return &check, nil
}

func (kc *KolideClient) GetDeviceFailure(deviceId int, failureId int) (*DeviceFailure, error) {
	var deviceFailures DeviceFailures
	cursor := ""

	apiUrl, err := url.Parse(kc.GetApiPathf("devices/%d/failures", deviceId))

	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}

	q := apiUrl.Query()
	q.Set("per_page", "2")
	apiUrl.RawQuery = q.Encode()

	for {
		log.Info(apiUrl.String())
		response, err := kc.client.Get(apiUrl.String())

		if err != nil {
			return nil, fmt.Errorf("getting devices failures: %w", err)
		}

		err = json.NewDecoder(response.Body).Decode(&deviceFailures)

		if err != nil {
			return nil, fmt.Errorf("decoding device failures: %w", err)
		}

		log.Infof("failures: %+v", deviceFailures)

		for _, deviceFailure := range deviceFailures.Data {
			if deviceFailure.Id == failureId {
				return &deviceFailure, nil
			}
		}

		cursor = deviceFailures.Pagination.NextCursor

		if cursor == "" {
			break
		}

		q.Set("cursor", deviceFailures.Pagination.NextCursor)
		apiUrl.RawQuery = q.Encode()
	}

	return nil, fmt.Errorf("failure with ID %d not found on device with ID %d", failureId, deviceId)
}
