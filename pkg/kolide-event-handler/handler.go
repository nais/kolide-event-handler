package kolide_event_handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/nais/kolide-event-handler/pkg/kolide"
	log "github.com/sirupsen/logrus"
)

func New(client *kolide.Client, events chan<- KolideEventFailure, signingSecret []byte) *KolideEventHandler {
	return &KolideEventHandler{
		signingSecret: signingSecret,
		events:        events,
		client:        client,
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
	requestBody, err := io.ReadAll(request.Body)
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

	log.Debugf("Kolide webhook triggered: %s", event.Event)

	switch event.Event {
	case "failures.new", "failures.resolved":
		var eventFailure KolideEventFailure

		err := json.Unmarshal(requestBody, &eventFailure)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		keh.events <- eventFailure

	case "webhook.test":
		log.Warnf("Kolide webhook test triggered with data '%s'", event.Event)

	default:
		log.Debugf("Unsupported event: %s", event.Event)
	}
}
