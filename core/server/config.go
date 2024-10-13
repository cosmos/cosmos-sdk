package server

// DynamicConfig defines an interface for configuration that can be dynamically
// fetched at runtime by an arbitrary key.
type DynamicConfig interface {
	Get(string) any
	GetString(string) string
}

type ConfigMap map[string]any

type ModuleConfigMap struct {
	Module string
	Config ConfigMap
}

func (ModuleConfigMap) IsManyPerContainerType() {}
