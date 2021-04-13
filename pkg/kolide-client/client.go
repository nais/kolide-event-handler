package kolide_client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func (kc *KolideClient) GetPaginated(path string, output interface{}) error {
	var data []interface{}
	cursor := ""

	apiUrl, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("create url: %w", err)
	}

	q := apiUrl.Query()
	q.Set("per_page", "100")
	apiUrl.RawQuery = q.Encode()

	for {
		log.Info(apiUrl.String())
		response, err := kc.client.Get(apiUrl.String())
		if err != nil {
			return fmt.Errorf("getting paginated response: %w", err)
		}

		responseBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("reading response body bytes: %w", err)
		}

		var paginatedResponse PaginatedResponse
		err = json.Unmarshal(responseBytes, &paginatedResponse)
		if err != nil {
			return fmt.Errorf("decoding paginated response: %w", err)
		}

		data = append(data, paginatedResponse.Data...)

		cursor = paginatedResponse.Pagination.NextCursor
		if cursor == "" {
			break
		}

		q.Set("cursor", paginatedResponse.Pagination.NextCursor)
		apiUrl.RawQuery = q.Encode()
	}

	// Here we reencode the data from map[string]interface{} to the correct type.
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshalling data interface: %w", err)
	}

	err = json.Unmarshal(b, output)
	if err != nil {
		return fmt.Errorf("unmarshalling data interface: %w", err)
	}

	return nil
}

func (kc *KolideClient) GetDeviceFailure(deviceId int, failureId int) (*DeviceFailure, error) {
	var deviceFailures []DeviceFailure

	err := kc.GetPaginated(kc.GetApiPathf("devices/%d/failures", deviceId), &deviceFailures)
	if err != nil {
		return nil, fmt.Errorf("getting device failures: %w", err)
	}

	for _, failure := range deviceFailures {
		if failure.Id == failureId {
			return &failure, nil
		}
	}

	return nil, fmt.Errorf("failure with ID %d not found on device with ID %d", failureId, deviceId)
}

func (kc *KolideClient) GetDevices() ([]Device, error) {
	var devices []Device

	err := kc.GetPaginated(kc.GetApiPath("/devices"), &devices)
	if err != nil {
		return nil, fmt.Errorf("getting devices: %w", err)
	}

	return devices, nil
}
