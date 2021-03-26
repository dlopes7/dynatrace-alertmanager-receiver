package communication

import (
	"fmt"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/alertmanager"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/cache"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/utils"
	dynatrace "github.com/dyladan/dynatrace-go-client/api"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type DynatraceController struct {
	customDeviceCache cache.CustomDeviceCacheService
	problemCache      cache.ProblemCacheService
	dtClient          dynatrace.Client
}

func NewDynatraceController() DynatraceController {
	dt := dynatrace.New(dynatrace.Config{
		APIKey:    os.Getenv("DT_API_KEY"),
		BaseURL:   os.Getenv("DT_BASE_URL"),
		Retries:   5,
		RetryTime: 2 * time.Second,
	})
	return DynatraceController{
		customDeviceCache: cache.NewCustomDeviceCacheService(),
		problemCache:      cache.NewProblemCacheService(),
		dtClient:          dt,
	}
}

func (d *DynatraceController) SendAlerts(data alertmanager.Data) error {

	// TODO - Implement Manual Closing the of problem - Need to get the ProblemID for the opened events
	// TODO - Check if I have the problem ID for each one of the sent events-
	// TODO - Resend events so that they don't expire in Dynatrace

	// Use the standard Custom Device Name for now, until we are able to build a new one from the labels of the alert
	// If we are not able to craft a new custom device name, this default name will be used
	customDeviceName := utils.CustomDeviceName
	eventProperties := map[string]string{
		"GroupKey": data.GroupKey,
	}
	eventType := dynatrace.EventType(dynatrace.EventTypeCustomInfo)
	description := fmt.Sprintf("Alert from AlertManager: %s", data.GroupKey)
	title := fmt.Sprintf("Alert from AlertManager")

	// This is our connection from this event to an eventual Problem in Dynatrace
	groupKeyHash := utils.Hash(data.GroupKey)

	log.WithFields(log.Fields{"groupKeyHash": groupKeyHash, "groupKey": data.GroupKey}).Info("Calculated the hash for the groupKey")

	for i, alert := range data.Alerts {
		alertIdentifier := fmt.Sprintf("Alert %d", i+1)
		log.WithFields(log.Fields{"alert": fmt.Sprintf("%+v", alert)}).Info("Processing alert")

		// Build the Custom Device name based on the namespace + service
		if namespace, ok := alert.Labels["namespace"]; ok {
			customDeviceName = fmt.Sprintf("Alertmanager - %s", namespace)
		}
		if service, ok := alert.Labels["service"]; ok {
			customDeviceName = fmt.Sprintf("%s: %s", customDeviceName, service)
		}

		// Change the title based on the alert name
		if alertname, ok := alert.Labels["alertname"]; ok {
			title = alertname
		}

		// Change the description based on the message
		if message, ok := alert.Annotations["message"]; ok {
			description = message
		}

		// Change the eventType of the Dynatrace Event based on the severity
		if severity, ok := alert.Labels["severity"]; ok {
			title = fmt.Sprintf("%s (%s)", title, severity)
			if severity == "warning" || severity == "error" || severity == "critical" {
				eventType = dynatrace.EventTypeErrorEvent
				title = fmt.Sprintf("%s (%s)", title, groupKeyHash)
			}
			log.WithFields(log.Fields{"severity": severity, "eventType": eventType}).Info("Setting eventType based on severity of the alert")
		}

		// Add labels and annotations as custom properties of the alert
		for key, value := range alert.Labels {
			propertyKey := fmt.Sprintf("%s - Label: %s", alertIdentifier, key)
			eventProperties[propertyKey] = value
		}
		for key, value := range alert.Annotations {
			propertyKey := fmt.Sprintf("%s - Annotation: %s", alertIdentifier, key)
			eventProperties[propertyKey] = value
		}
	}

	// Here we need to make sure we have a Custom Device before proceeding
	_, customDeviceID := utils.GenerateGroupAndCustomDeviceID(utils.CustomDeviceGroupName, customDeviceName)
	log.WithFields(log.Fields{"customDeviceID": customDeviceID, "customDeviceName": customDeviceName}).Info("Generated a Custom Device ID locally")
	customDeviceCache := d.customDeviceCache.GetCache()
	if !utils.StringInSlice(customDeviceID, customDeviceCache.CustomDevices) {
		// We don't have this Custom Device ID stored. We need to create a new Custom Device
		cd := dynatrace.CustomDevicePushMessage{
			DisplayName: customDeviceName,
			Group:       utils.CustomDeviceGroupName,
		}
		r, _, err := d.dtClient.CustomDevice.Create(customDeviceName, cd)
		if err != nil {
			// We were not able to create the custom device, abort
			return err
		}
		customDeviceCache.CustomDevices = append(customDeviceCache.CustomDevices, r.EntityID)
		d.customDeviceCache.Update(*customDeviceCache)
		log.WithFields(log.Fields{"CustomDeviceID": r.EntityID}).Info("Created a CustomDeviceID using the API")
	}

	// This means we need to send an event to Dynatrace
	if data.Status == "firing" {

		event := dynatrace.EventCreation{
			EventType:      eventType,
			Source:         "AlertManager",
			TimeoutMinutes: 120,
			AttachRules: dynatrace.PushEventAttachRules{
				EntityIds: []string{customDeviceID},
			},
			Description:      description,
			Title:            title,
			CustomProperties: eventProperties,
		}

		r, _, err := d.dtClient.Events.Create(event)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{"response": fmt.Sprintf("%+v", r)}).Info("Dynatrace response")

		if eventType == dynatrace.EventTypeErrorEvent {
			p := cache.Problem{
				Event: event,
				Alert: data,
			}
			d.problemCache.AddProblem(groupKeyHash, p)
		}
	}

	return nil
}
