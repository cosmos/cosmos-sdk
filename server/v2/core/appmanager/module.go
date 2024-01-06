package appmanager

import (
	"context"

	"cosmossdk.io/server/v2/core/transaction"
)

type Identity = []byte

type Tx = transaction.Tx

type MsgRouterBuilder interface {
	RegisterHandler(msg Type, handlerFunc func(ctx context.Context, msg Type) (resp Type, err error))
}

type QueryRouterBuilder = MsgRouterBuilder

type PreMsgRouterBuilder interface {
	RegisterPreHandler(msg Type, preHandler func(ctx context.Context, msg Type) error)
}

type PostMsgRouterBuilder interface {
	RegisterPostHandler(msg Type, postHandler func(ctx context.Context, msg, msgResp Type) error)
}

type STFModule[T transaction.Tx] interface {
	// TODO: should we separate the interface to avoid boilerplate in modules when things are not needed?
	Name() string
	RegisterMsgHandlers(router MsgRouterBuilder)
	RegisterQueryHandler(router QueryRouterBuilder)
	BeginBlocker() func(ctx context.Context) error
	EndBlocker() func(ctx context.Context) error
	UpdateValidators() func(ctx context.Context) ([]ValidatorUpdate, error)
	TxValidator() func(ctx context.Context, tx T) error // why does the module handle registration
	RegisterPreMsgHandler(router PreMsgRouterBuilder)
	RegisterPostMsgHandler(router PostMsgRouterBuilder)
}

type UpgradeModule interface {
	UpgradeBlocker() func(ctx context.Context) error
}

type Module[T transaction.Tx] interface {
	STFModule[T]
}

// Update defines what is expected to be returned
type ValidatorUpdate struct {
	PubKey []byte
	Power  int64 // updated power of the validtor
}
