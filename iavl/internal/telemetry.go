package internal

import (
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("iavl")
	meter  = otel.Meter("iavl")
	logger = otelslog.NewLogger("iavl")
)

func init() {
}
