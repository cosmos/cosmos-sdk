package appmodule

import (
	"reflect"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/core/internal"
)

// Register registers a module with the global module registry. The provided
// protobuf message is used only to identify uniquely the protobuf module config
// type. The instance of this message used in the configuration will be injected
// into the container and can be requested by a provider function. All module
// initialization should be handled by the provided options.
func Register(msg proto.Message, options ...Option) {
	ty := reflect.TypeOf(msg)
	init := &internal.ModuleInitializer{
		ConfigProtoMessage: msg,
		ConfigGoType:       ty,
	}
	internal.ModuleRegistry[ty] = init

	for _, option := range options {
		init.Error = option.apply(init)
		if init.Error != nil {
			return
		}
	}
}
