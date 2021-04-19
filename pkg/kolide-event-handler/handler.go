package kolide_event_handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
		signingSecret: signingSecret,
		apiClient:     kolideclient.New(apiToken),
		listChan:      listChan,
	}
}

func (keh *KolideEventHandler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/isalive", httpStatusOk)
	mux.HandleFunc("/isready", httpStatusOk)
	mux.HandleFunc("/webhooks", keh.handleWebhookEvent)

	return mux
}

func httpStatusOk(writer http.ResponseWriter, _ *http.Request) {
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
	keh.listChan <- &pb.DeviceList{
		// TODO get device
		Devices: []*pb.Device{{ }},
	}

	return nil
}

func (keh *KolideEventHandler) handleEventFailure(ctx context.Context, eventFailure KolideEventFailure) error {
	// look up severity for all checks this device currently fails on
	check, err := keh.apiClient.GetCheck(ctx, eventFailure.Data.CheckId)

	if err != nil {
		return fmt.Errorf("getting check: %w", err)
	}

	severity := GetSeverity(check)

	if severity < SeverityNotice {
		return nil
	}

	graceTime := GetGraceTime(severity)

	log.Infof("grace time: %v", graceTime)

	keh.listChan <- &pb.DeviceList{
		// TODO get device
		Devices: []*pb.Device{{ }},
	}

	return nil
}

func (keh *KolideEventHandler) Cron(programContext context.Context) {
	ticker := time.NewTicker(time.Second * 10)

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

			dl := &pb.DeviceList{
				Devices: devices,
			}

			keh.listChan <- dl
		case <-programContext.Done():
			log.Infof("Stoping cron")
			return
		}
	}
}