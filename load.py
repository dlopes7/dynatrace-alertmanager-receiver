import requests
import time

service = "allowDavisMerge_false_test_04"

warning = {
    "receiver": "dynatrace-receiver",
    "status": "firing",
    "alerts": [
        {
            "status": "firing",
            "labels": {
                "alertname": "TargetDown",
                "job": "kubelet",
                "namespace": "kube-system",
                "prometheus": "kubelet",
                "service": "kubelet" + service,
                "severity": "warning",
            },
            "annotations": {"message": "11.11% of the kubelet/kubelet targets in kube-system"},
            "startsAt": "2021-03-19T01:35:45.72Z",
            "endsAt": "0001-01-01T00:00:00Z",
            "generatorURL": "http://openshift.com",
            "fingerprint": "e425bb91067b6c9e",
        }
    ],
    "groupKey": '{}:{alertname="Test Alert", cluster="Cluster 02", service="Service ' + service,
    "groupLabels": {"alertname": "Test Alert", "cluster": "Cluster 02", "service": "Service 02"},
    "commonLabels": {"alertname": "Test Alert", "cluster": "Cluster 02", "service": "Service 02"},
    "commonAnnotations": {"annotation_01": "annotation 01", "annotation_02": "annotation 03"},
    "externalURL": "http://8598cebf58a1:9093",
}

info = {
    "receiver": "dynatrace-receiver",
    "status": "firing",
    "alerts": [
        {
            "status": "firing",
            "labels": {
                "alertname": "TargetDown",
                "job": "kubelet",
                "namespace": "kube-system",
                "prometheus": "kubelet",
                "service": "kubelet2",
                "severity": "info",
            },
            "annotations": {"message": "11.11% of the kubelet/kubelet targets in kube-system"},
            "startsAt": "2021-03-19T01:35:45.72Z",
            "endsAt": "0001-01-01T00:00:00Z",
            "generatorURL": "http://openshift.com",
            "fingerprint": "e425bb91067b6c9e",
        }
    ],
    "groupKey": '{}:{alertname="Test Alert", cluster="Cluster 02", service="Service 01"}',
    "groupLabels": {"alertname": "Test Alert", "cluster": "Cluster 02", "service": "Service 02"},
    "commonLabels": {"alertname": "Test Alert", "cluster": "Cluster 02", "service": "Service 02"},
    "commonAnnotations": {"annotation_01": "annotation 01", "annotation_02": "annotation 03"},
    "externalURL": "http://8598cebf58a1:9093",
}

warning_close = {
    "receiver": "dynatrace-receiver",
    "status": "resolved",
    "alerts": [
        {
            "status": "resolved",
            "labels": {
                "alertname": "TargetDown",
                "job": "kubelet",
                "namespace": "kube-system",
                "prometheus": "kubelet",
                "service": "kubelet" + service,
                "severity": "warning",
            },
            "annotations": {"message": "11.11% of the kubelet/kubelet targets in kube-system"},
            "startsAt": "2021-03-19T01:35:45.72Z",
            "endsAt": "0001-01-01T00:00:00Z",
            "generatorURL": "http://openshift.com",
            "fingerprint": "e425bb91067b6c9e",
        }
    ],
    "groupKey": '{}:{alertname="Test Alert", cluster="Cluster 02", service="Service ' + service,
    "groupLabels": {"alertname": "Test Alert", "cluster": "Cluster 02", "service": "Service 02"},
    "commonLabels": {"alertname": "Test Alert", "cluster": "Cluster 02", "service": "Service 02"},
    "commonAnnotations": {"annotation_01": "annotation 01", "annotation_02": "annotation 03"},
    "externalURL": "http://8598cebf58a1:9093",
}

def main():
    print(requests.post("http://localhost:9394/webhook", json=warning))
    # time.sleep(10)
    # print(requests.post("http://localhost:9393/webhook", json=warning_close))
    # time.sleep(10)
    # print(requests.post("http://localhost:9393/webhook", json=warning_close))


if __name__ == "__main__":
    main()
