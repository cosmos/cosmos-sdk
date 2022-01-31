package internal

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/container"
)

var ModuleRegistry = map[string]*ModuleInfo{}

type ModuleInfo struct {
	ConfigSection string
	ConfigType    reflect.Type
	Constructor   func(config interface{}) container.Option
}
