// Package registry provides an instrument registry for telemetry instrumentation.
package registry

import "fmt"

// Instrument defines an interface for telemetry instruments that can self-register.
type Instrument interface {
	// Name returns the instrument's identifier used in configuration.
	Name() string
	// Start initializes the instrument with the provided configuration map.
	Start(config map[string]any) error
}

var instruments = map[string]Instrument{}

// Register registers an instrument for use in configuration.
// This is typically called from an instrument package's init() function.
func Register(i Instrument) {
	if _, exists := instruments[i.Name()]; exists {
		panic(fmt.Sprintf("telemetry: instrument %q already registered", i.Name()))
	}
	instruments[i.Name()] = i
}

// Get returns a registered instrument by name, or nil if not found.
func Get(name string) Instrument {
	return instruments[name]
}
