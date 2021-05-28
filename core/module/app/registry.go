package app

import (
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/core/module"
)

var registry map[reflect.Type]Handler

func RegisterAppModule(constructor interface{}) {
	typ := reflect.TypeOf(constructor)
	if typ.Kind() != reflect.Func {
		panic("TODO")
	}

	if typ.NumIn() < 1 {
		panic("TODO")
	}

	typ.In(1)

	configField, ok := typ.FieldByName("Config")
	if !ok {
		panic(fmt.Errorf("module handler struct %T does not contain a Config field", handler))
	}

	if existing, ok := registry[configField.Type]; ok {
		panic(fmt.Errorf("module handler %T already registered for config type %T, trying to register new handler type %T", existing, configField.Type, handler))
	}

	registry[configField.Type] = handler
}

type ModuleSet struct {
	modMap map[string]Handler
}

func NewModuleSet(configMap module.ModuleConfigSet) (ModuleSet, error) {
	// TODO deterministic order
	modMap := make(map[string]Handler)
	for k, v := range configMap {
		mod, ok := registry[reflect.TypeOf(v)]
		if !ok {
			panic("TODO")
		}

		mod = mod.New()
		modMap[k] = mod
	}

	return ModuleSet{modMap: modMap}, nil
}
