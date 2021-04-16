package kolide_client

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
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

const MaxHttpRetries = 10
const DefaultRetryAfter = time.Second

func GetRetryAfter(header http.Header) time.Duration {
	limit := header.Get("Ratelimit-Limit")
	remaining := header.Get("Ratelimit-Remaining")
	reset := header.Get("Ratelimit-Reset")
	retryAfter := header.Get("Retry-After")

	if retryAfter == "" {
		return 0
	}

	log.Infof("rate-limited: limit: %s, remaining: %s, reset: %s, retry-after: %s", limit, remaining, reset, retryAfter)

	seconds, err := strconv.Atoi(retryAfter)

	if err != nil {
		retryAfterDate, err := time.Parse(time.RFC1123, retryAfter)
		if err != nil || retryAfterDate.Before(time.Now()) {
			return DefaultRetryAfter
		}

		return time.Until(retryAfterDate).Round(time.Second)
	}

	if seconds < 0 {
		return DefaultRetryAfter
	}

	return time.Second * time.Duration(seconds)
}

func (kc *KolideClient) Get(path string) (*http.Response, error) {
	for attempt := 0; attempt < MaxHttpRetries; attempt++ {
		resp, err := kc.client.Get(kc.GetApiPath(path))
		if err != nil {
			return nil, err
		}

		switch statusCode := resp.StatusCode; {
		case statusCode == http.StatusOK:
			return resp, nil
		case statusCode == http.StatusTooManyRequests:
			sleep := GetRetryAfter(resp.Header)
			log.Debugf("[attempt %d/%d] StatusTooManyRequests: sleeping %v", attempt, MaxHttpRetries, sleep)
			time.Sleep(sleep)
		case statusCode >= 500:
			sleep := time.Duration(attempt+1) * time.Second
			log.Debugf("[attempt %d/%d] KolideServerError: sleeping %v", attempt, MaxHttpRetries, sleep)
			time.Sleep(sleep)
		default:
			return nil, fmt.Errorf("unexpected stauts code: %d, response: %v", statusCode, resp)
		}
	}

	return nil, fmt.Errorf("max retries exceeded")
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

func (kc *KolideClient) GetDevices() ([]*Device, error) {
	var devices []*Device

	err := kc.GetPaginated(kc.GetApiPath("/devices"), &devices)
	if err != nil {
		return nil, fmt.Errorf("getting devices: %w", err)
	}

	err = kc.PopulateDevicesFailures(devices)
	if err != nil {
		log.Warnf("populating device failures: %v", err)
	}

	return devices, nil
}

func (kc *KolideClient) GetDeviceFailures(deviceId int) ([]*DeviceFailure, error) {
	var deviceFailures []*DeviceFailure
	err := kc.GetPaginated(kc.GetApiPathf("/devices/%d/failures", deviceId), &deviceFailures)
	if err != nil {
		return nil, fmt.Errorf("getting paginated device failures: %v", err)
	}

	return deviceFailures, nil
}

func (kc *KolideClient) PopulateDevicesFailures(devices []*Device) error {
	var multiError []error
	wg := sync.WaitGroup{}
	for _, device := range devices {
		if device.FailureCount == 0 {
			continue
		}

		wg.Add(1)
		go func(d *Device) {
			multiError = kc.PopulateDeviceFailures(d)
			wg.Done()
		}(device)
	}
	wg.Wait()

	if len(multiError) > 0 {
		for _, err := range multiError {
			log.Debugf("%v", err)
		}

		return fmt.Errorf("%d error(s) occurred while populating failures on devices", len(multiError))
	}

	return nil
}

func (kc *KolideClient) PopulateDeviceFailures(device *Device) []error {
	deviceFailures, err := kc.GetDeviceFailures(device.Id)
	if err != nil {
		return []error{fmt.Errorf("getting device failures: %w", err)}
	}

	wg := sync.WaitGroup{}
	var multiError []error
	for _, failure := range deviceFailures {
		wg.Add(1)
		go func(df *DeviceFailure) {
			err := kc.PopulateCheck(df)
			if err != nil {
				multiError = append(multiError, err)
			}
			wg.Done()
		}(failure)
	}
	wg.Wait()

	device.Failures = deviceFailures
	return multiError
}

func (kc *KolideClient) PopulateCheck(df *DeviceFailure) error {
	check, err := kc.GetCheck(df.CheckId)
	if err != nil {
		return fmt.Errorf("getting check: %w", err)
	}

	df.Check = check
	return nil
}

/*
devices
failures
checks
*/
