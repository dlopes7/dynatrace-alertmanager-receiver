module github.com/dlopes7/dynatrace-alertmanager-receiver

go 1.16

require (
	github.com/dyladan/dynatrace-go-client v1.0.1
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/prometheus/alertmanager v0.21.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.8.1
	github.com/twmb/murmur3 v1.1.5
	github.com/x-cray/logrus-prefixed-formatter v0.5.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)

replace github.com/dyladan/dynatrace-go-client => /home/david/projects/go/dynatrace-go-client