import requests
import time

from concurrent.futures import ThreadPoolExecutor

service = "-demo-05"

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
                # "label_code_app": "label_code_app",
                "label_env": "label_env",
                "ocp_cluster": "ocp4-intra-prod",
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

def make_request(url, body):
    try:
        print(requests.post(url, json=body))
    except Exception as e:
        print(e)


def main():
    make_request("http://localhost:9394/webhook", warning)


if __name__ == "__main__":
    main()
