package server

import (
	"encoding/json"
	"fmt"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/alertmanager"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/cache"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/dynatrace"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/jobs"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/utils"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Server struct {
	dt         dynatrace.DynatraceController
	problemJob jobs.ProblemJob
}

func New() Server {
	customDeviceCache := cache.NewCustomDeviceCacheService()
	problemCache := cache.NewProblemCacheService()
	problemJob := jobs.NewProblemJob(&customDeviceCache, &problemCache)

	return Server{
		dt:         dynatrace.NewDynatraceController(&customDeviceCache, &problemCache, &problemJob),
		problemJob: problemJob,
	}
}

func (s *Server) webhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	resp := utils.Response{}

	// Decode the incoming request body to a Data object
	data := alertmanager.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp = utils.Response{
			Error:   true,
			Message: fmt.Sprintf("Could not parse the from the request body: %s", err.Error()),
		}
		log.WithFields(log.Fields{"response": resp, "error": err.Error()}).Error("Server - Could not parse the data to a valid Data object")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	log.WithFields(log.Fields{"data": fmt.Sprintf("%+v", data)}).Info("Server - Received data")

	// Attempt to send the alerts to Dynatrace
	err := s.dt.SendAlerts(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp = utils.Response{
			Error:   true,
			Message: fmt.Sprintf("Could not send the alert to Dynatrace: %s", err.Error()),
		}
		log.WithFields(log.Fields{"response": resp, "error": err.Error()}).Error("Server - Could not send the alert to Dynatrace")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

}

func Run() {
	s := New()
	c := cron.New()
	c.AddFunc("@every 2m", s.problemJob.UpdateProblemIDs)
	c.Start()

	http.HandleFunc("/webhook", s.webhook)

	listenAddress := ":9393"
	if os.Getenv("WEBHOOK_PORT") != "" {
		listenAddress = ":" + os.Getenv("WEBHOOK_PORT")
	}

	log.WithFields(log.Fields{"listenAddress": listenAddress}).Info("Server - Starting webhook")
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
