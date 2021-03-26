module github.com/dlopes7/dynatrace-alertmanager-receiver

go 1.16

require (
	github.com/dyladan/dynatrace-go-client v1.0.0
	github.com/prometheus/alertmanager v0.21.0
	github.com/sirupsen/logrus v1.8.1
	github.com/twmb/murmur3 v1.1.5
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/dyladan/dynatrace-go-client => /home/david/projects/go/dynatrace-go-client
