package kolide

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Client struct {
	baseUrl string
	client  *http.Client
}

func New(token string) *Client {
	t := Transport{
		Token: token,
	}

	return &Client{
		baseUrl: "https://k2.kolide.com/api/v0",
		client:  t.Client(),
	}
}

const (
	MaxHttpRetries    = 10
	DefaultRetryAfter = time.Second
)

func GetRetryAfter(header http.Header) time.Duration {
	limit := header.Get("Ratelimit-Limit")
	remaining := header.Get("Ratelimit-Remaining")
	reset := header.Get("Ratelimit-Reset")
	retryAfter := header.Get("Retry-After")

	if retryAfter == "" {
		return 0
	}

	log.Debugf("rate-limited: limit: %s, remaining: %s, reset: %s, retry-after: %s", limit, remaining, reset, retryAfter)

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

func respectTheirAuthority(sleep time.Duration) {
	n := rand.Intn(10)
	time.Sleep(sleep + (time.Second * time.Duration(n)))
}

func (kc *Client) get(ctx context.Context, path string) (*http.Response, error) {
	multiError := []error{}
	for attempt := 0; attempt < MaxHttpRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("making http request: %w", err)
		}

		resp, err := kc.client.Do(req)
		if err != nil {
			multiError = append(multiError, err)
			log.Debugf("[attempt %d/%d] Error: %v.", attempt, MaxHttpRetries, err)
			continue
		}

		switch statusCode := resp.StatusCode; {
		case statusCode == http.StatusOK:
			return resp, nil
		case statusCode == http.StatusTooManyRequests:
			sleep := GetRetryAfter(resp.Header)
			log.Debugf("[attempt %d/%d] StatusTooManyRequests: sleeping %v", attempt, MaxHttpRetries, sleep)
			respectTheirAuthority(sleep)
		case statusCode >= 500:
			sleep := time.Duration(attempt+1) * time.Second
			log.Debugf("[attempt %d/%d] KolideServerError: sleeping %v", attempt, MaxHttpRetries, sleep)
			respectTheirAuthority(sleep)
		default:
			message, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status code: %d, response: %v", statusCode, string(message))
		}
	}

	return nil, fmt.Errorf("max retries exceeded: errors: %w", errors.Join(multiError...))
}

func (kc *Client) GetApiPath(path string) string {
	return fmt.Sprintf("%s/%s", kc.baseUrl, path)
}

func (kc *Client) GetApiPathf(path string, args ...interface{}) string {
	return kc.GetApiPath(fmt.Sprintf(path, args...))
}

func (kc *Client) GetDevice(ctx context.Context, deviceId uint64) (*Device, error) {
	response, err := kc.get(ctx, kc.GetApiPathf("devices/%d", deviceId))
	if err != nil {
		return nil, fmt.Errorf("getting client: %w", err)
	}

	defer response.Body.Close()

	var device Device
	err = json.NewDecoder(response.Body).Decode(&device)
	if err != nil {
		return nil, fmt.Errorf("decoding device: %w", err)
	}

	multiError := kc.PopulateDeviceFailures(ctx, &device)

	if len(multiError) > 0 {
		log.Debugf("multiple errors occurred: %+v", multiError)

		return nil, fmt.Errorf("%d error(s) occurred while populating failures on devices", len(multiError))
	}

	return &device, nil
}

func (kc *Client) GetCheck(ctx context.Context, checkId int) (*Check, error) {
	response, err := kc.get(ctx, kc.GetApiPathf("checks/%d", checkId))
	if err != nil {
		return nil, fmt.Errorf("getting check: %w", err)
	}

	defer response.Body.Close()

	var check Check
	err = json.NewDecoder(response.Body).Decode(&check)
	if err != nil {
		return nil, fmt.Errorf("decoding check: %w", err)
	}

	return &check, nil
}

func (kc *Client) GetPaginated(ctx context.Context, path string) ([]json.RawMessage, error) {
	var data []json.RawMessage
	cursor := ""

	apiUrl, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("create url: %w", err)
	}

	q := apiUrl.Query()
	q.Set("per_page", "100")
	apiUrl.RawQuery = q.Encode()

	for {
		response, err := kc.get(ctx, apiUrl.String())
		if err != nil {
			return nil, fmt.Errorf("getting paginated response: %w", err)
		}

		//goland:noinspection GoDeferInLoop
		defer response.Body.Close()

		responseBytes, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body bytes: %w", err)
		}

		var paginatedResponse PaginatedResponse
		err = json.Unmarshal(responseBytes, &paginatedResponse)
		if err != nil {
			return nil, fmt.Errorf("decoding paginated response: %w", err)
		}

		data = append(data, paginatedResponse.Data...)

		cursor = paginatedResponse.Pagination.NextCursor
		if cursor == "" {
			break
		}

		q.Set("cursor", paginatedResponse.Pagination.NextCursor)
		apiUrl.RawQuery = q.Encode()
	}

	return data, nil
}

func (kc *Client) GetDeviceFailure(ctx context.Context, deviceId int, failureId int) (*DeviceFailure, error) {
	failures, err := kc.GetDeviceFailures(ctx, deviceId)
	if err != nil {
		return nil, fmt.Errorf("get device failures: %w", err)
	}

	for _, failure := range failures {
		if failure.Id == failureId {
			return failure, nil
		}
	}

	return nil, fmt.Errorf("failure with ID %d not found on device with ID %d", failureId, deviceId)
}

func (kc *Client) GetDevices(ctx context.Context) ([]*Device, error) {
	log.Debugf("Getting all devices...")
	rawDevices, err := kc.GetPaginated(ctx, kc.GetApiPath("/devices"))
	if err != nil {
		return nil, err
	}

	var devices []*Device
	for _, rawDevice := range rawDevices {
		device := &Device{}
		err := json.Unmarshal(rawDevice, device)
		if err != nil {
			return nil, fmt.Errorf("unmarshal device: %w", err)
		}

		devices = append(devices, device)
	}

	err = kc.PopulateDevicesFailures(ctx, devices)
	if err != nil {
		log.Warnf("populating device failures: %v", err)
	}

	return devices, nil
}

func (kc *Client) GetDeviceFailures(ctx context.Context, deviceId int) ([]*DeviceFailure, error) {
	rawFailures, err := kc.GetPaginated(ctx, kc.GetApiPathf("/devices/%d/failures", deviceId))
	if err != nil {
		return nil, fmt.Errorf("getting paginated device failures: %v", err)
	}

	var failures []*DeviceFailure
	for _, rawFailure := range rawFailures {
		failure := &DeviceFailure{}
		err := json.Unmarshal(rawFailure, failure)
		if err != nil {
			return nil, fmt.Errorf("unmarshal failure: %w", err)
		}

		failures = append(failures, failure)
	}

	return failures, nil
}

func (kc *Client) PopulateDevicesFailures(ctx context.Context, devices []*Device) error {
	var multiError []error
	wg := sync.WaitGroup{}

	log.Debugf("Getting failure states for all devices...")

	for _, device := range devices {
		if device.FailureCount == 0 {
			continue
		}

		wg.Add(1)
		go func(d *Device) {
			multiError = kc.PopulateDeviceFailures(ctx, d)
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

func (kc *Client) PopulateDeviceFailures(ctx context.Context, device *Device) []error {
	deviceFailures, err := kc.GetDeviceFailures(ctx, device.Id)
	if err != nil {
		return []error{fmt.Errorf("getting device failures: %w", err)}
	}

	wg := sync.WaitGroup{}
	var multiError []error
	for _, failure := range deviceFailures {
		wg.Add(1)
		go func(df *DeviceFailure) {
			err := kc.PopulateCheck(ctx, df)
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

func (kc *Client) PopulateCheck(ctx context.Context, df *DeviceFailure) error {
	check, err := kc.GetCheck(ctx, df.CheckId)
	if err != nil {
		return fmt.Errorf("getting check: %w", err)
	}

	df.Check = check
	return nil
}
