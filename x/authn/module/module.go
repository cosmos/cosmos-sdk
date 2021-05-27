package module

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/core/module/app"

	"github.com/cosmos/cosmos-sdk/core/module"
	"github.com/cosmos/cosmos-sdk/x/authn"
)

func init() {
	module.RegisterModuleHandler(handler{})
}

type handler struct {
	*authn.Module
}

func (h handler) ConfigType() proto.Message {
	return h.Module
}

func (h handler) New(config proto.Message) module.ModuleHandler {
	mod := config.(*authn.Module)
	return handler{mod}
}

type AppModuleDeps struct {
	Key app.RootModuleKey
}

func (h handler) NewAppModule(deps AppModuleDeps) app.Module {
	panic("TODO")
}
