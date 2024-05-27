package v2

import "fmt"

func (x *ModuleConfig) GetTypeUrl() (string, error) {
	if x != nil {
		if x.Config != nil {
			return x.Config.TypeUrl, nil
		} else {
			return "", fmt.Errorf("module %q is missing a config object", x.Name)
		}
	}
	return "", fmt.Errorf("Module is nil")
}

func (x *ModuleConfig) GetGolangBindingsStrings() ([]string, []string) {
	if x != nil {
		interfaceTypes := make([]string, len(x.GolangBindings))
		imple := make([]string, len(x.GolangBindings))

		for i, binding := range x.GolangBindings {
			interfaceTypes[i] = binding.InterfaceType
			imple[i] = binding.Implementation
		}
		return interfaceTypes, imple
	}
	return nil, nil
}