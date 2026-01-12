package internal

import "go.opentelemetry.io/otel"
import "go.opentelemetry.io/contrib/bridges/otelslog"

var (
	Tracer = otel.Tracer("iavl")
	Meter  = otel.Meter("iavl")
	Logger = otelslog.NewLogger("iavl")
)
