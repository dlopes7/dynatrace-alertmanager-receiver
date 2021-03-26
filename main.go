package main

import (
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/server"
	log "github.com/sirupsen/logrus"
	"os"
)

func init() {

	log.SetLevel(log.DebugLevel)
	logFormatter := &log.TextFormatter{
		DisableColors:   true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05.000",
	}
	log.SetFormatter(logFormatter)

	//logDir := utils.GetTempDir()
	//logFilePath := path.Join(logDir, "dynatrace-receiver.log")
	//lumberjackLogger := &lumberjack.Logger{
	//	Filename:   logFilePath,
	//	MaxSize:    5,
	//	MaxBackups: 5,
	//}
	// log.SetOutput(lumberjackLogger)
	log.SetOutput(os.Stdout)

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
