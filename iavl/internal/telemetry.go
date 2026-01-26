package internal

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	// TODO these shouldn't be exported from internal package
	Tracer           = otel.Tracer("iavl")
	Meter            = otel.Meter("iavl")
	LeafHashLatency  metric.Int64Histogram
	WALWritelLatency metric.Int64Histogram
)

func init() {
	var err error

	LeafHashLatency, err = Meter.Int64Histogram("iavl_leaf_hash_latency_ms",
		metric.WithDescription("The amount of time in commit needed to wait for leaf hashing to complete, before hashing the root"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_leaf_hash_latency_ms histogram: " + err.Error())
	}

	WALWritelLatency, err = Meter.Int64Histogram("iavl_wal_write_latency_ms",
		metric.WithDescription("The amount of time we had to wait in Commit before returning, due to WAL writes"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_wal_write_latency_ms histogram: " + err.Error())
	}
}
