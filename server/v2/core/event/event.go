<<<<<<< HEAD
package event

// Events represents a list of events.
type Events struct {
	Events []Event
}

// Event defines how an event will emitted
type Event struct {
	Type       string
	Attributes []Attribute
}

// KVEventAttribute is a kv-pair event attribute.
type Attribute struct {
	Key, Value string
}
||||||| 39865d852f
=======
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
>>>>>>> marko/app_manager
