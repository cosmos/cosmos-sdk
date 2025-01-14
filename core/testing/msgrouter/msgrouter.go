package msgrouter

import (
	"context"
	"fmt"

	gogoproto "github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/core/router"
	"cosmossdk.io/core/transaction"
)

// msgTypeURL returns the TypeURL of a proto message.
func msgTypeURL(msg gogoproto.Message) string {
	return gogoproto.MessageName(msg)
}

type routerHandler func(context.Context, transaction.Msg) (transaction.Msg, error)

var _ router.Service = &RouterService{}

// custom router service for integration tests
type RouterService struct {
	handlers map[string]routerHandler
}

func NewRouterService() *RouterService {
	return &RouterService{
		handlers: make(map[string]routerHandler),
	}
}

func (rs *RouterService) RegisterHandler(handler routerHandler, typeUrl string) {
	rs.handlers[typeUrl] = handler
}

func (rs RouterService) CanInvoke(ctx context.Context, typeUrl string) error {
	if rs.handlers[typeUrl] == nil {
		return fmt.Errorf("no handler for typeURL %s", typeUrl)
	}
	return nil
}

func (rs RouterService) Invoke(ctx context.Context, req transaction.Msg) (transaction.Msg, error) {
	typeUrl := msgTypeURL(req)
	if rs.handlers[typeUrl] == nil {
		return nil, fmt.Errorf("no handler for typeURL %s", typeUrl)
	}
	return rs.handlers[typeUrl](ctx, req)
}
