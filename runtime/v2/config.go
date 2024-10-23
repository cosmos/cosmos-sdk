package runtime

import (
	"cosmossdk.io/core/server"
	"cosmossdk.io/depinject"
	"strings"
)

// GlobalConfig is a recursive configuration map containing configuration
// key-value pairs parsed from the configuration file, flags, or other
// input sources.
type GlobalConfig server.ConfigMap

// ModuleConfigMaps is a map module scoped ConfigMaps
type ModuleConfigMaps map[string]server.ConfigMap

// TODO combine below 2 functions
// - linear search for module name in provider is OK
// - move elsewhere, server/v2 or runtime/v2 ?

// ProvideModuleConfigMaps returns a map of module name to module config map.
// The module config map is a map of flag to value.
func ProvideModuleConfigMaps(
	moduleConfigs []server.ModuleConfigMap,
	globalConfig GlobalConfig,
) ModuleConfigMaps {
	moduleConfigMaps := make(ModuleConfigMaps)
	for _, moduleConfig := range moduleConfigs {
		cfg := moduleConfig.Config
		name := moduleConfig.Module
		moduleConfigMaps[name] = make(server.ConfigMap)
		for flag, df := range cfg {
			m := globalConfig
			fetchFlag := flag
			for _, part := range strings.Split(flag, ".") {
				if maybeMap, ok := m[part]; ok {
					innerMap, ok := maybeMap.(map[string]any)
					if !ok {
						fetchFlag = part
						break
					}
					m = innerMap
				} else {
					break
				}
			}
			if val, ok := m[fetchFlag]; ok {
				moduleConfigMaps[name][flag] = val
			} else {
				moduleConfigMaps[name][flag] = df
			}
		}
	}
	return moduleConfigMaps
}

func ProvideModuleScopedConfigMap(
	key depinject.ModuleKey,
	moduleConfigs ModuleConfigMaps,
) server.ConfigMap {
	return moduleConfigs[key.Name()]
}
