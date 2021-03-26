# Dynatrace Alertmanager Receiver

This is an [Alertmanager](https://prometheus.io/docs/alerting/latest/alertmanager/) receiver webhook for Dynatrace

### Features

* Sends Alertmanager alerts to different Custom Devices
* Creates Custom Devices based on labels, if available
* Sends custom info and problem opening events
* Automatically closes Dynatrace Problems when the alerts are resolved
* Periodically retrieve the Problem ID of sent events
* Periodically deletes stale events
* Periodically resends events to keep them opened in Dynatrace

### Environment Variables

* `DT_API_KEY` - The dynatrace API Key, mandatory
* `DT_BASE_URL` - The dynatrace API URL, mandatory
* `WEBHOOK_LOG_FOLDER` - The temp folder for logs and caches, if empty `os.TempDir()` is used.
* `WEBHOOK_PORT` - The webhook port, if empty `9393` is used
