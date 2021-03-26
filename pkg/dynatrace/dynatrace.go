package dynatrace

import (
	"fmt"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/alertmanager"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/cache"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/jobs"
	"github.com/dlopes7/dynatrace-alertmanager-receiver/pkg/utils"
	dtapi "github.com/dyladan/dynatrace-go-client/api"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

type DynatraceController struct {
	customDeviceCache *cache.CustomDeviceCacheService
	problemCache      *cache.ProblemCacheService
	problemJob        *jobs.ProblemJob
	dtClient          dtapi.Client
}

func NewDynatraceController(deviceCache *cache.CustomDeviceCacheService, problemCache *cache.ProblemCacheService, problemJob *jobs.ProblemJob) DynatraceController {
	dt := dtapi.New(dtapi.Config{
		APIKey:    os.Getenv("DT_API_KEY"),
		BaseURL:   os.Getenv("DT_BASE_URL"),
		Retries:   5,
		RetryTime: 2 * time.Second,
	})
	return DynatraceController{
		dtClient:          dt,
		customDeviceCache: deviceCache,
		problemCache:      problemCache,
	}
}

func (d *DynatraceController) SendAlerts(data alertmanager.Data) error {

	// TODO - Implement Resend Events Job
	// TODO - Implement Delete Old Events Job

	// Use the standard Custom Device Name for now, until we are able to build a new one from the labels of the alert
	// If we are not able to craft a new custom device name, this default name will be used
	customDeviceName := utils.CustomDeviceName
	eventProperties := map[string]string{
		"GroupKey": data.GroupKey,
	}
	eventType := dtapi.EventType(dtapi.EventTypeCustomInfo)
	description := fmt.Sprintf("Alert from AlertManager: %s", data.GroupKey)
	title := fmt.Sprintf("Alert from AlertManager")

	// This is our connection from this event to an eventual Problem in Dynatrace
	groupKeyHash := utils.Hash(data.GroupKey)

	log.WithFields(log.Fields{"groupKeyHash": groupKeyHash, "groupKey": data.GroupKey}).Info("Controller - Calculated the hash for the groupKey")

	for i, alert := range data.Alerts {
		alertIdentifier := fmt.Sprintf("Alert %d", i+1)
		log.WithFields(log.Fields{"alert": fmt.Sprintf("%+v", alert)}).Info("Controller - Processing alert")

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
				eventType = dtapi.EventTypeErrorEvent
				title = fmt.Sprintf("%s (%s)", title, groupKeyHash)
			}
			log.WithFields(log.Fields{"severity": severity, "eventType": eventType}).Info("Controller - Setting eventType based on severity of the alert")
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
	log.WithFields(log.Fields{"customDeviceID": customDeviceID, "customDeviceName": customDeviceName}).Info("Controller - Generated a Custom Device ID locally")

	// This means we need to send an event to Dynatrace
	if data.Status == "firing" {

		// Before sending an event, make sure the Custom Device exists
		customDeviceCache := d.customDeviceCache.GetCache()
		if !utils.StringInSlice(customDeviceID, customDeviceCache.CustomDevices) {
			// We don't have this Custom Device ID stored. We need to create a new Custom Device
			cd := dtapi.CustomDevicePushMessage{
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
			log.WithFields(log.Fields{"CustomDeviceID": r.EntityID}).Info("Controller - Created a new Custom Device using the API")
		} else {
			log.WithFields(log.Fields{"CustomDeviceID": customDeviceID}).Info("Controller - Found the CustomDeviceID in the local cache")
		}

		// Create the event object
		event := dtapi.EventCreation{
			EventType:      eventType,
			Source:         "AlertManager",
			TimeoutMinutes: 120,
			AttachRules: dtapi.PushEventAttachRules{
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

		log.WithFields(log.Fields{"response": fmt.Sprintf("%+v", r)}).Info("Controller - Dynatrace response")

		if eventType == dtapi.EventTypeErrorEvent {
			p := cache.Problem{
				Event: event,
				Alert: data,
			}
			d.problemCache.AddProblem(groupKeyHash, p)
		}
	} else if data.Status == "resolved" && eventType == dtapi.EventTypeErrorEvent {
		// If we get here, we need to manually close the Dynatrace Problem

		log.WithFields(log.Fields{"groupKeyHash": groupKeyHash}).Info("Controller - Received a resolved error event, need to close the problem")
		err := d.CloseProblem(groupKeyHash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DynatraceController) CloseProblem(groupKeyHash string) error {
	comment := fmt.Sprintf("Dynatrace receiver automatically closed the problem after receiving an resolved event with hash %s", groupKeyHash)
	problemCache := d.problemCache.GetCache()
	if cachedProblem, ok := problemCache.Problems[groupKeyHash]; ok {
		if cachedProblem.ProblemID != "" {
			log.WithFields(log.Fields{"hash": groupKeyHash, "problem": cachedProblem.ProblemID}).Info("Controller - Found problem, closing it")
			_, err := d.dtClient.Problem.Close(cachedProblem.ProblemID, comment)
			if err != nil {
				return err
			}
		} else {
			// We could not find the ProblemID for this event, maybe it resolved too fast, before the Problem Job could have updated it
			log.WithFields(log.Fields{"groupKeyHash": groupKeyHash}).Warning("Controller - Found an event on the ProblemCache, but no ProblemID, attempting to update the cache now")
			d.problemJob.UpdateProblemIDs()

			// Get an updated cache, after the job manual run
			problemCache = d.problemCache.GetCache()
			if cachedProblem, ok := problemCache.Problems[groupKeyHash]; ok {
				if cachedProblem.ProblemID != "" {
					_, err := d.dtClient.Problem.Close(cachedProblem.ProblemID, comment)
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("found the event (%s) in the cache, but could not get a ProblemID from Dynatrace even after a manual scan", groupKeyHash)
				}
			}
		}

	} else {
		d.problemCache.Delete(groupKeyHash)
		return fmt.Errorf("could not find an event with hash %s in the ProblemCache, can't close the event", groupKeyHash)
	}

	// If we get here, the problem has been closed
	log.WithFields(log.Fields{"hash": groupKeyHash}).Info("Controller - The problem has been closed successfully")
	d.problemCache.Delete(groupKeyHash)
	return nil

}
