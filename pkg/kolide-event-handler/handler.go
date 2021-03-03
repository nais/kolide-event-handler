package kolide_event_handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	kolide_client "github.com/nais/kolide-event-handler/pkg/kolide-client"
	log "github.com/sirupsen/logrus"
)

func New(signingSecret []byte, apiToken string) *KolideEventHandler {
	return &KolideEventHandler{
		signingSecret: signingSecret,
		apiClient:     kolide_client.New(apiToken),
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

		err = keh.handleEventFailure(eventFailure)

		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			log.Warnf("Event handling: %v", err)
			return
		}
	default:
		log.Infof("Unsupported event: %s", event.Event)
	}
}

func (keh *KolideEventHandler) handleEventFailure(eventFailure KolideEventFailure) error {
	check, err := keh.apiClient.GetCheck(eventFailure.Data.CheckId)

	if err != nil {
		return fmt.Errorf("getting check: %w", err)
	}

	severity := GetSeverity(*check)

	if severity < SEVERITY_NOTICE {
		return nil
	}

	graceTime := GetGraceTime(severity)

	log.Infof("grace time: %v", graceTime)

	return nil
}

func GetSeverity(check kolide_client.Check) Severity {
	var severity, mostSevereSeverity Severity = -1, -1

	for _, tag := range check.Tags {
		switch strings.ToLower(tag) {
		case "info":
			severity = SEVERITY_INFO
		case "notice":
			severity = SEVERITY_NOTICE
		case "warning":
			severity = SEVERITY_WARNING
		case "danger":
			severity = SEVERITY_DANGER
		case "critical":
			severity = SEVERITY_CRITICAL
		}

		if severity > mostSevereSeverity {
			mostSevereSeverity = severity
		}
	}

	if mostSevereSeverity == -1 {
		log.Warnf("Check missing a severity tag: %+v", check)
		mostSevereSeverity = SEVERITY_WARNING
	}

	return mostSevereSeverity
}

func GetGraceTime(severity Severity) time.Duration {
	switch severity {
	case SEVERITY_NOTICE:
		return DURATION_NOTICE
	case SEVERITY_WARNING:
		return DURATION_WARNING
	case SEVERITY_DANGER:
		return DURATION_DANGER
	case SEVERITY_CRITICAL:
		return DURATION_CRITICAL
	default:
		log.Errorf("Unknown severity: %v", severity)
		return DURATION_UNKNOWN
	}
}
