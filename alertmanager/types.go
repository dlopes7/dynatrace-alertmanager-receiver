package alertmanager

import "github.com/prometheus/alertmanager/template"

type Data struct {
	Receiver string          `json:"receiver"`
	Status   string          `json:"status"`
	Alerts   template.Alerts `json:"alerts"`
	GroupKey string          `json:"groupKey"`

	GroupLabels       template.KV `json:"groupLabels"`
	CommonLabels      template.KV `json:"commonLabels"`
	CommonAnnotations template.KV `json:"commonAnnotations"`

	ExternalURL string `json:"externalURL"`
}
