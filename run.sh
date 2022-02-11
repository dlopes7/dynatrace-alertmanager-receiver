#/bin/bash

export DT_API_TOKEN=DRLsfmuKScmmIQSxtuzxJ
export DT_API_URL=https://eaa50379.sprint.dynatracelabs.com/
export WEBHOOK_PORT=9394
export WEBHOOK_LOG_LEVEL=DEBUG
export WEBHOOK_PROBLEM_SEVERITIES=critical,warning,error

go run main.go
