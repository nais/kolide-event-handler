package kolide_event_handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/ioutil"
	"net/http"
	"time"

	kolideclient "github.com/nais/kolide-event-handler/pkg/kolide-client"
	"github.com/nais/kolide-event-handler/pkg/pb"

	log "github.com/sirupsen/logrus"
)

const FullSyncInterval = time.Minute * 5

func New(listChan chan<- *pb.DeviceList, signingSecret []byte, apiToken string) *KolideEventHandler {
	return &KolideEventHandler{
		signingSecret:        signingSecret,
		apiClient:            kolideclient.New(apiToken),
		deviceListUpdateChan: listChan,
	}
}

func (keh *KolideEventHandler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/isalive", httpStatusOk)
	mux.HandleFunc("/isready", httpStatusOk)
	mux.HandleFunc("/webhooks", keh.handleWebhookEvent)

	return mux
}

func httpStatusOk(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (keh *KolideEventHandler) handleWebhookEvent(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mac := hmac.New(sha256.New, keh.signingSecret)
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		log.Warnf("Request body: %v", err)
		return
	}

	log.Tracef("Request body: %s", string(requestBody))
	mac.Write(requestBody)

	incomingSignature, err := hex.DecodeString(request.Header.Get("Authorization"))
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	expectedSignature := mac.Sum(nil)
	if !hmac.Equal(incomingSignature, expectedSignature) {
		writer.WriteHeader(http.StatusForbidden)
		return
	}

	var event KolideEvent
	err = json.Unmarshal(requestBody, &event)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
	}

	switch event.Event {
	case "failures.new", "failures.resolved":
		var eventFailure KolideEventFailure

		err := json.Unmarshal(requestBody, &eventFailure)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = keh.handleEventFailure(request.Context(), eventFailure)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Warnf("Event handling: %v", err)
			return
		}
	case "webhook.test":
		_ = keh.handleEventTest(event)
	default:
		log.Infof("Unsupported event: %s", event.Event)
	}
}

func (keh *KolideEventHandler) handleEventTest(event KolideEvent) error {
	log.Infof("got test event: %+v", event)
	keh.deviceListUpdateChan <- &pb.DeviceList{
		Devices: []*pb.Device{{
			Id:     uint64(133769420),
			Serial: "testserial",
		}},
	}

	return nil
}

func (keh *KolideEventHandler) handleEventFailure(ctx context.Context, eventFailure KolideEventFailure) error {
	device, err := keh.apiClient.GetDevice(ctx, eventFailure.Data.DeviceId)
	if err != nil {
		return fmt.Errorf("get device from api: %w", err)
	}

	keh.deviceListUpdateChan <- &pb.DeviceList{
		Devices: []*pb.Device{
			KolideDeviceToProtobufDevice(device),
		},
	}

	return nil
}

func (keh *KolideEventHandler) Cron(programContext context.Context) {
	ticker := time.NewTicker(time.Second * 1)

	for {
		select {
		case <-ticker.C:
			ticker.Reset(FullSyncInterval)
			log.Info("Doing full Kolide device health sync")

			ctx, cancel := context.WithTimeout(programContext, time.Minute)
			devices, err := keh.apiClient.GetDevices(ctx)
			cancel()
			if err != nil {
				log.Errorf("getting devies: %v", err)
			}

			keh.deviceListUpdateChan <- &pb.DeviceList{
				Devices: KolideDevicesToProtobufDevices(devices),
			}
		case <-programContext.Done():
			log.Infof("Stoping cron")
			return
		}
	}
}

func KolideFailuredToProtobufFailures(failures []*kolideclient.DeviceFailure) []*pb.Failure {
	var pbfailures []*pb.Failure

	for _, failure := range failures {
		pbfailures = append(pbfailures, &pb.Failure{
			Id:         uint64(failure.Id),
			Title:      failure.Title,
			CheckId:    uint64(failure.CheckId),
			Timestamp:  timestamppb.New(failure.Timestamp),
			ResolvedAt: timestamppb.New(failure.ResolvedAt),
			Ignored:    failure.Ignored,
			Check: &pb.Check{
				Tags: failure.Check.Tags,
			},
		})
	}

	return pbfailures
}

func KolideDevicesToProtobufDevices(devices []*kolideclient.Device) []*pb.Device {
	var pbdevices []*pb.Device

	for _, device := range devices {
		pbdevices = append(pbdevices, KolideDeviceToProtobufDevice(device))
	}

	return pbdevices
}

func KolideDeviceToProtobufDevice(device *kolideclient.Device) *pb.Device {
	return &pb.Device{
		Id:              uint64(device.Id),
		Name:            device.Name,
		OwnedBy:         device.OwnedBy,
		Platform:        device.Platform,
		LastSeenAt:      timestamppb.New(device.LastSeenAt),
		FailureCount:    uint64(device.FailureCount),
		PrimaryUserName: device.PrimaryUserName,
		Serial:          device.Serial,
		AssignedOwner: &pb.Owner{
			Email: device.AssignedOwner.Email,
		},
		Failures: KolideFailuredToProtobufFailures(device.Failures),
	}
}
