package main

import (
	"net/http"
	"os"
	"time"

	keh "github.com/nais/kolide-event-handler/pkg/kolide-event-handler"
	log "github.com/sirupsen/logrus"
)

func main() {
	keh := keh.New([]byte(os.Getenv("KOLIDE_SIGNING_SECRET")))
	mux := keh.Routes()

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatalf("Serving: %v", err)
	}
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
