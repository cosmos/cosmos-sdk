package module

import (
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/gogo/protobuf/proto"
)

type ModuleHandler interface {
	// ModuleType returns the configuration type for this module
	ConfigType() proto.Message

	// New returns a new module handler
	New(config proto.Message) ModuleHandler
}

var registry map[reflect.Type]ModuleHandler

func RegisterModuleHandler(handler ModuleHandler) {
	typ := reflect.TypeOf(handler.ConfigType())
	if _, ok := registry[typ]; ok {
		panic(fmt.Errorf("module handler for config type %T already registered", handler.ConfigType()))
	}

	registry[typ] = handler
}

type ModuleSet struct {
	handlers map[string]ModuleHandler
}

func NewModuleSet(moduleConfigs map[string]*types.Any) *ModuleSet {
	panic("TODO")
}

func (ms *ModuleSet) Each(f func(name string, handler ModuleHandler)) {
	panic("TODO")
}
