package telemetry

type Metrics interface {
	Gather(format string) (GatherResponse, error)
}

var _ Metrics = &metrics{}
