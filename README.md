# Dynatrace Alertmanager Receiver

This is an alertmanager receiver webhook for Dynatrace

### Features

* Sends Alertmanager alerts to different Custom Devices
* Creates Custom Devices based on labels, if available
* Sends custom info and problem opening events
* Automatically closes Dynatrace Problems when the alerts are resolved
* Periodically retrieve the Problem ID of sent events
* Periodically deletes stale events
* Periodically resends events to keep them opened in Dynatrace

