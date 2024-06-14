package webhook_handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type KolideEventHandler struct {
	signingSecret []byte
	events        chan<- KolideEventFailure

	log logrus.FieldLogger
}

type KolideEventFailureData struct {
	CheckId   int    `json:"check_id"`
	FailureId int    `json:"failure_id"`
	Title     string `json:"title"`
	DeviceId  int    `json:"device_id"`
}

type KolideEventFailure struct {
	Id        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      KolideEventFailureData `json:"data"`
	KolideEvent
}

type KolideEvent struct {
	Event string `json:"event"`
}

func New(events chan<- KolideEventFailure, signingSecret []byte, log logrus.FieldLogger) *KolideEventHandler {
	return &KolideEventHandler{
		signingSecret: signingSecret,
		events:        events,
		log:           log,
	}
}

func (w *KolideEventHandler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/isalive", httpStatusOk)
	mux.HandleFunc("/isready", httpStatusOk)
	mux.HandleFunc("/webhooks", w.handleWebhookEvent)

	return mux
}

func httpStatusOk(writer http.ResponseWriter, _ *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

func (w *KolideEventHandler) handleWebhookEvent(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mac := hmac.New(sha256.New, w.signingSecret)
	requestBody, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		w.log.Warnf("Request body: %v", err)
		return
	}

	w.log.Debugf("Request body: %s", string(requestBody))
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

	w.log.Debugf("Kolide webhook triggered: %s", event.Event)

	switch event.Event {
	case "failures.new", "failures.resolved":
		var eventFailure KolideEventFailure

		err := json.Unmarshal(requestBody, &eventFailure)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		w.events <- eventFailure

	case "webhook.test":
		w.log.Warnf("Kolide webhook test triggered with data '%s'", event.Event)

	default:
		w.log.Debugf("Unsupported event: %s", event.Event)
	}
}
