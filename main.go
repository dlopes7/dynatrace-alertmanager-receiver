package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	dynatrace "github.com/dyladan/dynatrace-go-client/api"
	"github.com/prometheus/alertmanager/template"
	log "github.com/sirupsen/logrus"
	"github.com/twmb/murmur3"
	"gopkg.in/natefinch/lumberjack.v2"
	"net/http"
	"os"
	"path"
	"strconv"
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

func GenerateGroupAndCustomDeviceID(groupID string, deviceID string) (string, string) {

	namespace := ""
	namespaceLenBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint32(namespaceLenBigEndian, uint32(len(namespace)))

	groupLenBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint32(groupLenBigEndian, uint32(len(groupLenBigEndian)))

	var groupIdBytes []byte
	groupIdBytes = append(groupIdBytes, namespaceLenBigEndian[:len(namespaceLenBigEndian)-4]...)
	groupIdBytes = append(groupIdBytes, []byte(groupID)...)
	groupIdBytes = append(groupIdBytes, groupLenBigEndian[:len(groupLenBigEndian)-4]...)

	dtGroupID := dtMurMur3(groupIdBytes)
	dtGroupIDUint64, _ := strconv.ParseUint(dtGroupID, 16, 64)

	dtGroupIDBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint64(dtGroupIDBigEndian, dtGroupIDUint64)

	deviceIDLenBigEndian := make([]byte, 8)
	binary.BigEndian.PutUint32(deviceIDLenBigEndian, uint32(len(deviceID)))

	var customDeviceBytes []byte
	customDeviceBytes = append(customDeviceBytes, dtGroupIDBigEndian...)
	customDeviceBytes = append(customDeviceBytes, []byte(deviceID)...)
	customDeviceBytes = append(customDeviceBytes, deviceIDLenBigEndian[:len(deviceIDLenBigEndian)-4]...)

	dtCustomDeviceID := dtMurMur3(customDeviceBytes)

	return fmt.Sprintf("CUSTOM_DEVICE_GROUP-%s", dtGroupID), fmt.Sprintf("CUSTOM_DEVICE-%s", dtCustomDeviceID)
}

func dtMurMur3(data []byte) string {

	h1, _ := murmur3.SeedSum128(0, 0, data)
	return fmt.Sprintf("%X", h1)

}
