package kolide_event_handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func New(signingSecret []byte) *KolideEventHandler {
	return &KolideEventHandler{
		signingSecret: signingSecret,
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

	log.Infof("Request body: %s", string(requestBody))

	mac.Write(requestBody)

	incomingSignature, err := hex.DecodeString(request.Header.Get("Authorization"))

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
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

		err = handleEventFailure(eventFailure)

		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Warnf("Event handling: %v", err)
			return
		}
	default:
		log.Infof("Unsupported event: %s", event.Event)
	}
}

func handleEventFailure(eventFailure KolideEventFailure) error {
	return nil
}
