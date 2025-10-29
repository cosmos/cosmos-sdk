package telemetry

import (
	"cosmossdk.io/log"
	"github.com/hashicorp/go-metrics"
)

type LoggerSink struct {
	logger log.Logger
}

func NewLoggerSink(logger log.Logger) *LoggerSink {
	return &LoggerSink{logger: logger}
}

func (l *LoggerSink) logWithLabels(labels []metrics.Label, kvs ...any) {
	logger := l.logger
	if len(labels) != 0 {
		kvs := make([]any, 0, len(labels)*2)
		for _, label := range labels {
			kvs = append(kvs, label.Name, label.Value)
		}
		logger = logger.With(kvs...)
	}
	logger.Info("telemetry", kvs...)
}

func (l *LoggerSink) SetGauge(key []string, val float32) {
	l.SetGaugeWithLabels(key, val, nil)
}

func (l *LoggerSink) SetGaugeWithLabels(key []string, val float32, labels []metrics.Label) {
	l.logWithLabels(labels, "type", "gauge", "key", key, "val", val)
}

func (l *LoggerSink) EmitKey(key []string, val float32) {
	l.logWithLabels(nil, "type", "kv", "key", key, "val", val)
}

func (l *LoggerSink) IncrCounter(key []string, val float32) {
	l.IncrCounterWithLabels(key, val, nil)
}

func (l *LoggerSink) IncrCounterWithLabels(key []string, val float32, labels []metrics.Label) {
	l.logWithLabels(labels, "type", "incr_counter", "key", key, "val", val)
}

func (l *LoggerSink) AddSample(key []string, val float32) {
	l.AddSampleWithLabels(key, val, nil)
}

func (l *LoggerSink) AddSampleWithLabels(key []string, val float32, labels []metrics.Label) {
	l.logWithLabels(labels, "type", "add_sample", "key", key, "val", val)
}

var _ metrics.MetricSink = (*LoggerSink)(nil)
