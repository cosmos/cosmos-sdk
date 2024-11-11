package server

import "strings"

// DynamicConfig defines an interface for configuration that can be dynamically
// fetched at runtime by an arbitrary key.
type DynamicConfig interface {
	Get(string) any
	GetString(string) string
}

// ConfigMap is a recursive map of configuration values.
type ConfigMap map[string]any

func (c ConfigMap) getSubConfig(key string) (ConfigMap, []string) {
	parts := strings.Split(key, ".")
	subConfig := c
	for i, part := range parts {
		if val, ok := subConfig[part]; ok {
			if nestedConfig, ok := val.(map[string]any); ok {
				subConfig = nestedConfig
			} else {
				return nil, parts[i:]
			}
		} else {
			return nil, parts[i:]
		}
	}
	return subConfig, nil
}

func getValue[T any](cfg ConfigMap, key string, def T) any {
	cfg, remaining := cfg.getSubConfig(key)
	if cfg == nil || len(remaining) > 0 {
		return def
	}
	if len(remaining) == 0 {
		return cfg
	}
	res, ok := cfg[remaining[0]]
	if !ok {
		return def
	}
	return res
}

// Get implements DynamicConfig.
func (c ConfigMap) Get(key string) any {
	return getValue[any](c, key, nil)
}

// GetInt64 implements DynamicConfig.
func (c ConfigMap) GetInt64(key string) int64 {
	return getValue(c, key, int64(0)).(int64)
}

// GetSliceOfStringSlices implements DynamicConfig.
func (c ConfigMap) GetSliceOfStringSlices(key string) [][]string {
	return getValue(c, key, [][]string{}).([][]string)
}

// GetString implements DynamicConfig.
func (c ConfigMap) GetString(key string) string {
	return getValue(c, key, "").(string)
}

// GetUint64 implements DynamicConfig.
func (c ConfigMap) GetUint64(key string) uint64 {
	return getValue(c, key, uint64(0)).(uint64)
}

var _ DynamicConfig = ConfigMap{}

// ModuleConfigMap is used to specify module configuration.
// Keys (and there default values and types) should be set in Config
// and returned by module specific provider function.
type ModuleConfigMap struct {
	Module string
	Config ConfigMap
}

func (ModuleConfigMap) IsManyPerContainerType() {}
