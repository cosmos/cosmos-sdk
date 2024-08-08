package router

import (
	"context"

	gogoproto "github.com/cosmos/gogoproto/proto"
)

// Service is the interface that wraps the basic methods for a router.
// A router can be a query router or a message router.
type Service interface {
	// CanInvoke returns an error if the given request cannot be invoked.
	CanInvoke(ctx context.Context, typeURL string) error
	// Invoke execute a message or query. It should be used when the called doesn't know the type of the response.
	Invoke(ctx context.Context, req gogoproto.Message) (res gogoproto.Message, err error)
}
