package main

import (
	"encoding/json"
	"fmt"
	dynatrace "github.com/dyladan/dynatrace-go-client/api"
	"github.com/prometheus/alertmanager/template"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"path"
	"time"
)

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

var dtClient dynatrace.Client
var logger *log.Logger

func init() {

	logger = log.New()
	logger.SetLevel(log.DebugLevel)
	logFormatter := &log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
	logger.SetFormatter(logFormatter)

	tmpDir := os.TempDir()
	if os.Getenv("WEBHOOK_LOG_FOLDER") != "" {
		tmpDir = os.Getenv("WEBHOOK_LOG_FOLDER")
	}
	logDir := fmt.Sprintf("%s/dynatrace-receiver", tmpDir)

	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		logger.WithFields(log.Fields{"path": logDir}).Info("Creating temporary logs directory")
		err := os.MkdirAll(tmpDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	logFilePath := path.Join(logDir, "dynatrace-receiver.log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5,
		MaxBackups: 5,
	}
	logger.SetOutput(lumberjackLogger)

	dtClient = dynatrace.New(dynatrace.Config{
		APIKey:  os.Getenv("DT_API_KEY"),
		BaseURL: os.Getenv("DT_BASE_URL"),
		Log:     logger,
	})

}

func sendAlertsToDynatrace(data template.Data) error {
	for _, alert := range data.Alerts {
		logger.WithFields(log.Fields{"alert": fmt.Sprintf("%+v", alert)}).Info("Processing alert")

		event := dynatrace.EventCreation{
			EventType: "ERROR_EVENT",
			Start:     alert.StartsAt.UnixNano() / int64(time.Millisecond),
			Source:    "AlertManager",
			AttachRules: dynatrace.PushEventAttachRules{
				EntityIds: []string{"CUSTOM_DEVICE-D3692A3DBB1B6419"},
			},
			Description:      fmt.Sprintf("Alert from AlertManager: %s", alert.Status),
			Title:            fmt.Sprintf("Alert from AlertManager"),
			CustomProperties: alert.Labels,
		}

		r, _, err := dtClient.Events.Create(event)
		if err != nil {
			return err
		}
		logger.WithFields(log.Fields{"response": fmt.Sprintf("%+v", r)}).Info("Dynatrace response")

	}
	return nil
}

func webhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	resp := Response{}

	// Decode the incoming request body to a template.Data object
	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp = Response{
			Error:   true,
			Message: fmt.Sprintf("Could not parse the from the request body: %s", err.Error()),
		}
		logger.WithFields(log.Fields{"response": resp, "error": err.Error()}).Error("Could not parse the data to a valid Data object")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	logger.WithFields(log.Fields{"data": fmt.Sprintf("%+v", data)}).Info("Received data")

	// Attempt to send the alerts to Dynatrace
	err := sendAlertsToDynatrace(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp = Response{
			Error:   true,
			Message: fmt.Sprintf("Could not send the alert to Dynatrace: %s", err.Error()),
		}
		logger.WithFields(log.Fields{"response": resp, "error": err.Error()}).Error("Could not send the alert to Dynatrace")
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

}

func main() {

	http.HandleFunc("/webhook", webhook)

	listenAddress := ":9393"
	if os.Getenv("WEBHOOK_PORT") != "" {
		listenAddress = ":" + os.Getenv("WEBHOOK_PORT")
	}

	logger.WithFields(log.Fields{"listenAddress": listenAddress}).Info("Starting webhook")
	logger.Fatal(http.ListenAndServe(listenAddress, nil))

}

/*
curl -i \
  'http://localhost:9093/api/v2/alerts' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '[
  {
    "startsAt": "2021-03-18T23:27:45.720Z",
    "annotations": {
      "annotation_01": "annotation 01",
      "annotation_02": "annotation 02",
      "annotation_02": "annotation 03"
    },
    "labels": {
      "alertname": "Test Alert",
      "cluster": "Cluster 01",
      "service": "Service 03"
    },
    "generatorURL": "http://openshift.com"
  }
]'
*/
