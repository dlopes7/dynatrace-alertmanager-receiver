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

* `DT_API_TOKEN` - The dynatrace API Key, mandatory
* `DT_API_URL` - The dynatrace API URL, mandatory
* `DT_GROUP_NAME` - The dynatrace Group Name
* `WEBHOOK_LOG_FOLDER` - The temp folder for logs and caches, if empty `os.TempDir()` is used.
* `WEBHOOK_PORT` - The webhook port, if empty `9393` is used
* `WEBHOOK_LOG_LEVEL` - The log level, if empty `INFO` is used
* `WEBHOOK_PROBLEM_SEVERITIES` - Comma separated of severities that open problems, ie: `critical,warning,error`

### Example curl to test

```bash
curl -i 'http://172.28.16.1:9393/webhook' \
-d '{
   "receiver":"dynatrace-receiver",
   "status":"firing",
   "alerts":[
      {
         "status":"firing",
         "labels":{
            "alertname":"TargetDown",
            "job":"kubelet",
            "namespace":"kube-system",
            "prometheus":"kubelet",
            "service":"kubelet",
            "severity":"warning"
         },
         "annotations":{
            "message":"11.11% of the kubelet/kubelet targets in kube-system"
         },
         "startsAt":"2021-03-19T01:35:45.72Z",
         "endsAt":"0001-01-01T00:00:00Z",
         "generatorURL":"http://openshift.com",
         "fingerprint":"e425bb91067b6c9e"
      }
   ],
   "groupKey":"{}:{\"alertname\": \"Test Alert\", \"cluster\": \"Cluster 02\", \"service\": \"Service 01\"}",
   "groupLabels":{
      "alertname":"Test Alert",
      "cluster":"Cluster 02",
      "service":"Service 02"
   },
   "commonLabels":{
      "alertname":"Test Alert",
      "cluster":"Cluster 02",
      "service":"Service 02"
   },
   "commonAnnotations":{
      "annotation_01":"annotation 01",
      "annotation_02":"annotation 03"
   },
   "externalURL":"http://8598cebf58a1:9093"
}'
```


