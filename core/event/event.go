package event

// Attribute is a kv-pair event attribute.
type Attribute struct {
	Key, Value string
}

func NewAttribute(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Events represents a list of events.
type Events struct {
	Events []Event
}

// Event defines how an event will emitted
type Event struct {
	Type       string
	Attributes []Attribute
}
