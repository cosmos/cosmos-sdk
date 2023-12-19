package event

import (
	"cosmossdk.io/core/event"
)

// Events represents a list of events.
type Events struct {
	Events []Event
}

// Event defines how an event will emitted
type Event struct {
	Type       string
	Attributes []event.Attribute
}
