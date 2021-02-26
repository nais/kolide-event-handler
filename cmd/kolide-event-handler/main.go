package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type KolideEventHandler struct {
	signingSecret []byte
}

type KolideEventFailureData struct {
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

func main() {
	keh := &KolideEventHandler{
		signingSecret: []byte(os.Getenv("KOLIDE_SIGNING_SECRET")),
	}

	mux := keh.Setup()
	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatalf("Serving: %v", err)
	}
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

func (keh *KolideEventHandler) Setup() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/isalive", httpStatusOk)
	mux.HandleFunc("/isready", httpStatusOk)
	mux.HandleFunc("/webhooks", keh.handleWebhookEvent)

	return mux
}

func cron() {
	ticker := time.NewTicker(time.Second * 1)
	for {
		select {
		case <-ticker.C:
			log.Info("Doing full Kolide device health sync")
		}
	}
}
