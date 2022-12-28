package metrics

import "github.com/armon/go-metrics"

func NewLabel(name, value string) metrics.Label {
	return metrics.Label{Name: name, Value: value}
}
