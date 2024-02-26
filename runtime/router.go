package runtime

import (
	"cosmossdk.io/core/router"
	"cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/gogoproto/proto"
)

func NewMsgRouterService(storeService store.KVStoreService, router baseapp.MessageRouter) router.Service {
	return &msgRouterService{
		storeService: storeService,
		router:       router,
	}
}

type msgRouterService struct {
	storeService store.KVStoreService
	router       baseapp.MessageRouter
}

// Handler implements router.Service.
func (m *msgRouterService) Handler(msg proto.Message) router.Service {
	return m.router.Handler(msg)
}

// HandlerByTypeURL implements router.Service.
func (*msgRouterService) HandlerByTypeURL(typeURL string) router.Service {
	panic("unimplemented")
}
