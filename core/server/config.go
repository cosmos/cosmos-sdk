package server

// DynamicConfig defines an interface for configuration that can be dynamically
// fetched at runtime by an arbitrary key.
type DynamicConfig interface {
	Get(string) any
	GetString(string) string
}

// ConfigMap is a recursive map of configuration values.
type ConfigMap map[string]any

// ModuleConfigMap is used to specify module configuration.
// Keys (and there default values and types) should be set in Config
// and returned by module specific provider function.
type ModuleConfigMap struct {
	Module string
	Config ConfigMap
}

func (ModuleConfigMap) IsManyPerContainerType() {}
