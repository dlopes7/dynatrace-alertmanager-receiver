package utils

import (
	dynatrace "github.com/dyladan/dynatrace-go-client/api"
)

const CustomDeviceGroupName = "Alertmanager"
const CustomDeviceName = "Alertmanager Events"

type Alert struct {
	Problem Problem                 `json:"problem"`
	Event   dynatrace.EventCreation `json:"event"`
}

type Problem struct {
	ProblemID string `json:"problemID"`
}

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}
