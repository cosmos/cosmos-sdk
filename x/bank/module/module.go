package module

import (
	"github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func init() {
	//module.RegisterModuleHandler(handler{})
}

type handler struct {
	*types.Module
}

func (h handler) ConfigType() proto.Message {
	return h.Module
}

//func (h handler) New(config proto.Message) module.ModuleHandler {
//	mod := config.(*types.Module)
//	return handler{mod}
//}
//
//type AppModuleDeps struct {
//	Key              app.RootModuleKey
//	AuthnQueryClient authn.QueryClient
//	AuthnMsgClient   authn.MsgClient
//}
//
//func (h handler) NewAppModule(deps AppModuleDeps) app.Handler {
//	panic("TODO")
//}
