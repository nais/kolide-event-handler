package kolide_event_handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/nais/kolide-event-handler/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func New(deviceEventChan chan<- *pb.DeviceEvent, signingSecret []byte) *KolideEventHandler {
	return &KolideEventHandler{
		signingSecret:     signingSecret,
		notifyDeviceEvent: deviceEventChan,
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

	log.Infof("got event: %s", event.Event)
	log.Debugf("event: %s", event.Event)
	switch event.Event {
	case "failures.new", "failures.resolved":
		var eventFailure KolideEventFailure

		err := json.Unmarshal(requestBody, &eventFailure)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		keh.handleEventFailure(eventFailure)
	case "webhook.test":
		keh.handleEventTest(event)
	default:
		log.Infof("Unsupported event: %s", event.Event)
	}
}

func (keh *KolideEventHandler) handleEventTest(event KolideEvent) {
	keh.notifyDeviceEvent <- &pb.DeviceEvent{
		Id:        "testid",
		DeviceId:  1,
		CheckId:   2,
		FailureId: 3,
		Event:     event.Event,
		Title:     "test title",
		Timestamp: timestamppb.Now(),
	}
}

func (keh *KolideEventHandler) handleEventFailure(eventFailure KolideEventFailure) {
	keh.notifyDeviceEvent <- &pb.DeviceEvent{
		Id:        eventFailure.Id,
		DeviceId:  uint64(eventFailure.Data.DeviceId),
		CheckId:   uint64(eventFailure.Data.CheckId),
		FailureId: uint64(eventFailure.Data.FailureId),
		Event:     eventFailure.Event,
		Title:     eventFailure.Data.Title,
		Timestamp: timestamppb.New(eventFailure.Timestamp),
	}
}
