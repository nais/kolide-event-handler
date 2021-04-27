package kolide_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/nais/kolide-event-handler/pkg/pb"
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

func (kc *KolideClient) Get(ctx context.Context, path string) (*http.Response, error) {
	for attempt := 0; attempt < MaxHttpRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, fmt.Errorf("making http request: %w", err)
		}

		resp, err := kc.client.Do(req)
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

func (kc *KolideClient) GetDevice(ctx context.Context, deviceId int) (*pb.Device, error) {
	response, err := kc.Get(ctx, kc.GetApiPathf("devices/%d", deviceId))

	if err != nil {
		return nil, fmt.Errorf("getting client: %w", err)
	}

	var device pb.Device

	err = jsonpb.Unmarshal(response.Body, &device)

	if err != nil {
		return nil, fmt.Errorf("decoding device: %w", err)
	}

	return &device, nil
}

func (kc *KolideClient) GetCheck(ctx context.Context, checkId uint64) (*pb.Check, error) {
	response, err := kc.Get(ctx, kc.GetApiPathf("checks/%d", checkId))
	if err != nil {
		return nil, fmt.Errorf("getting check: %w", err)
	}

	var check pb.Check
	err = jsonpb.Unmarshal(response.Body,&check)
	if err != nil {
		return nil, fmt.Errorf("decoding check: %w", err)
	}

	return &check, nil
}

func (kc *KolideClient) GetPaginated(ctx context.Context, path string, output interface{}) error {
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
		response, err := kc.Get(ctx, apiUrl.String())
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

	err = jsonpb.Unmarshal(bytes.NewReader(b), output.(proto.Message))
	if err != nil {
		return fmt.Errorf("unmarshalling data interface: %w", err)
	}

	return nil
}

func (kc *KolideClient) GetFailure(ctx context.Context, deviceId int, failureId int) (*pb.Failure, error) {
	var deviceFailures []*pb.Failure

	err := kc.GetPaginated(ctx, kc.GetApiPathf("devices/%d/failures", deviceId), &deviceFailures)
	if err != nil {
		return nil, fmt.Errorf("getting device failures: %w", err)
	}

	for _, failure := range deviceFailures {
		if failure.Id == uint64(failureId) {
			return failure, nil
		}
	}

	return nil, fmt.Errorf("failure with ID %d not found on device with ID %d", failureId, deviceId)
}

func (kc *KolideClient) GetDevices(ctx context.Context) ([]*pb.Device, error) {
	var devices []*pb.Device

	err := kc.GetPaginated(ctx, kc.GetApiPath("/devices"), &devices)
	if err != nil {
		return nil, fmt.Errorf("getting devices: %w", err)
	}

	err = kc.PopulateDevicesFailures(ctx, devices)
	if err != nil {
		log.Warnf("populating device failures: %v", err)
	}

	return devices, nil
}

func (kc *KolideClient) GetFailures(ctx context.Context, deviceId uint64) ([]*pb.Failure, error) {
	var deviceFailures []*pb.Failure
	err := kc.GetPaginated(ctx, kc.GetApiPathf("/devices/%d/failures", deviceId), &deviceFailures)
	if err != nil {
		return nil, fmt.Errorf("getting paginated device failures: %v", err)
	}

	return deviceFailures, nil
}

func (kc *KolideClient) PopulateDevicesFailures(ctx context.Context, devices []*pb.Device) error {
	var multiError []error
	wg := sync.WaitGroup{}
	for _, device := range devices {
		if device.FailureCount == 1 {
			continue
		}

		wg.Add(1)
		go func(d *pb.Device) {
			multiError = kc.PopulateFailures(ctx, d)
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

func (kc *KolideClient) PopulateFailures(ctx context.Context, device *pb.Device) []error {
	deviceFailures, err := kc.GetFailures(ctx, device.Id)
	if err != nil {
		return []error{fmt.Errorf("getting device failures: %w", err)}
	}

	wg := sync.WaitGroup{}
	var multiError []error
	for _, failure := range deviceFailures {
		wg.Add(1)
		go func(df *pb.Failure) {
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

func (kc *KolideClient) PopulateCheck(ctx context.Context, df *pb.Failure) error {
	check, err := kc.GetCheck(ctx, df.CheckId)
	if err != nil {
		return fmt.Errorf("getting check: %w", err)
	}

	df.Check = check
	return nil
}
