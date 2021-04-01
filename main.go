package main

import (
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/server"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/utils"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
)

func init() {

	log.SetLevel(log.InfoLevel)
	customLevel := os.Getenv("WEBHOOK_LOG_LEVEL")
	if customLevel != "" {
		level, err := log.ParseLevel(customLevel)
		if err != nil {
			log.Fatalf("Could not use level %s: %s", customLevel, err.Error())
		}
		log.SetLevel(level)
	}

	logFormatter := &prefixed.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		ForceFormatting: true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
	log.SetFormatter(logFormatter)

	logFilePath := path.Join(utils.GetTempDir(), "dynatrace-receiver.log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    5,
		MaxBackups: 5,
	}

	w := io.MultiWriter(os.Stdout, lumberjackLogger)
	log.SetOutput(w)

}

func main() {
	server.Run()
}

/*
curl -i 'http://localhost:9093/api/v2/alerts' \
-H 'accept: application/json' \
-H 'Content-Type: application/json' \
-d '[  {
    "startsAt": "2021-03-19T01:35:45.720Z",
    "annotations": {
      "annotation_01": "annotation 01",
      "annotation_02": "annotation 02",
      "annotation_02": "annotation 03"
    },
    "labels": {
      "alertname": "Test Alert",
      "cluster": "Cluster 01",
      "service": "Service 01"
    },
    "generatorURL": "http://openshift.com"
  }
]'

*/

/*
docker run --rm \
-e "DT_API_TOKEN=DRLsfmuKScmmIQSxtuzxJ" \
-e "DT_API_URL=https://eaa50379.sprint.dynatracelabs.com/" \
-e "WEBHOOK_PORT=9394" \
-e "WEBHOOK_LOG_LEVEL=DEBUG" \
-e "WEBHOOK_PROBLEM_SEVERITIES=critical,warning,error" \
-e "WEBHOOK_LOG_FOLDER=/tmp/webhook" \
--name dynatrace-receiver \
dlopes7/dynatrace-prometheus-receiver
*/
