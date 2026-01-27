package iavl

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	// TODO these shouldn't be exported from internal package
	tracer          = otel.Tracer("iavl")
	meter           = otel.Meter("iavl")
	logger          = otelslog.NewLogger("iavl")
	leafHashLatency metric.Int64Histogram
	walWriteLatency metric.Int64Histogram
)

func init() {
	var err error

	leafHashLatency, err = meter.Int64Histogram("iavl_leaf_hash_latency_ms",
		metric.WithDescription("The amount of time in commit needed to wait for leaf hashing to complete, before hashing the root"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_leaf_hash_latency_ms histogram: " + err.Error())
	}

	walWriteLatency, err = meter.Int64Histogram("iavl_wal_write_latency_ms",
		metric.WithDescription("The amount of time we had to wait in Commit before returning, due to WAL writes"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		panic("failed to create iavl_wal_write_latency_ms histogram: " + err.Error())
	}
}
