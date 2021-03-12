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

	kolideclient "github.com/nais/kolide-event-handler/pkg/kolide-client"
	"github.com/nais/kolide-event-handler/pkg/pb"

	log "github.com/sirupsen/logrus"
)

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

	// look up severity for all checks this device currently fails on

	check, err := keh.apiClient.GetCheck(eventFailure.Data.CheckId)

	if err != nil {
		return fmt.Errorf("getting check: %w", err)
	}

	severity := GetSeverity(*check)

	if severity < SeverityNotice {
		return nil
	}

	graceTime := GetGraceTime(severity)

	log.Infof("grace time: %v", graceTime)

	keh.listChan <- &pb.DeviceList{
		Devices: []*pb.DeviceHealthEvent{{
			DeviceId: uint64(eventFailure.Data.DeviceId),
			Health:   false,
			LastSeen: nil,
			Serial:   "",
			Username: "",
		}},
	}

	return nil
}

func GetSeverity(check kolideclient.Check) Severity {
	var severity, mostSevereSeverity Severity = -1, -1

	for _, tag := range check.Tags {
		switch strings.ToLower(tag) {
		case "info":
			severity = SeverityInfo
		case "notice":
			severity = SeverityNotice
		case "warning":
			severity = SeverityWarning
		case "danger":
			severity = SeverityDanger
		case "critical":
			severity = SeverityCritical
		}

		if severity > mostSevereSeverity {
			mostSevereSeverity = severity
		}
	}

	if mostSevereSeverity == -1 {
		log.Warnf("Check missing a severity tag: %+v", check)
		mostSevereSeverity = SeverityWarning
	}

	return mostSevereSeverity
}

func GetGraceTime(severity Severity) time.Duration {
	switch severity {
	case SeverityNotice:
		return DurationNotice
	case SeverityWarning:
		return DurationWarning
	case SeverityDanger:
		return DurationDanger
	case SeverityCritical:
		return DurationCritical
	default:
		log.Errorf("Unknown severity: %v", severity)
		return DurationUnknown
	}
}
