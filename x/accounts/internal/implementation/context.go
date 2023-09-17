package implementation

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/prefixstore"
)

var errUnauthorized = errors.New("unauthorized")

type contextKey struct{}

type contextValue struct {
	store             store.KVStore   // store is the prefixed store for the account.
	sender            []byte          // sender is the address of the entity invoking the account action.
	whoami            []byte          // whoami is the address of the account being invoked.
	originalContext   context.Context // originalContext that was used to build the account context.
	getExpectedSender func(msg proto.Message) ([]byte, error)
	moduleExec        func(ctx context.Context, msg proto.Message) (proto.Message, error)
	moduleQuery       func(ctx context.Context, msg proto.Message) (proto.Message, error)
}

// MakeAccountContext creates a new account execution context given:
// storeSvc: which fetches the x/accounts module store.
// accountAddr: the address of the account being invoked, which is used to give the
// account a prefixed storage.
// sender: the address of entity invoking the account action.
func MakeAccountContext(
	ctx context.Context,
	storeSvc store.KVStoreService,
	accountAddr,
	sender []byte,
	getSenderFunc func(msg proto.Message) ([]byte, error),
	moduleExec func(ctx context.Context, msg proto.Message) (proto.Message, error),
	moduleQuery func(ctx context.Context, msg proto.Message) (proto.Message, error),
) context.Context {
	return context.WithValue(ctx, contextKey{}, contextValue{
		store:             prefixstore.New(storeSvc.OpenKVStore(ctx), accountAddr),
		sender:            sender,
		whoami:            accountAddr,
		originalContext:   ctx,
		getExpectedSender: getSenderFunc,
		moduleExec:        moduleExec,
		moduleQuery:       moduleQuery,
	})
}

// ExecModule can be used to execute a message towards a module.
func ExecModule[Resp any, RespProto ProtoMsg[Resp], Req any, ReqProto ProtoMsg[Req]](ctx context.Context, msg ReqProto) (RespProto, error) {
	// get sender
	v := ctx.Value(contextKey{}).(contextValue)
	// check sender
	expectedSender, err := v.getExpectedSender(msg)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(expectedSender, v.whoami) {
		return nil, errUnauthorized
	}

	// execute module, unwrapping the original context.
	resp, err := v.moduleExec(v.originalContext, msg)
	if err != nil {
		return nil, err
	}

	concreteResp, ok := resp.(RespProto)
	if !ok {
		return nil, fmt.Errorf("unexpected response type %T", resp)
	}
	return concreteResp, nil
}

// QueryModule can be used by an account to execute a module query.
func QueryModule[Resp any, RespProto ProtoMsg[Resp], Req any, ReqProto ProtoMsg[Req]](ctx context.Context, msg ReqProto) (RespProto, error) {
	// we do not need to check the sender in a query because it is not a state transition.
	// we also unwrap the original context.
	v := ctx.Value(contextKey{}).(contextValue)
	resp, err := v.moduleQuery(v.originalContext, msg)
	if err != nil {
		return nil, err
	}
	concreteResp, ok := resp.(RespProto)
	if !ok {
		return nil, fmt.Errorf("unexpected response type %T", resp)
	}
	return concreteResp, nil
}

// OpenKVStore returns the prefixed store for the account given the context.
func OpenKVStore(ctx context.Context) store.KVStore {
	return ctx.Value(contextKey{}).(contextValue).store
}

// Sender returns the address of the entity invoking the account action.
func Sender(ctx context.Context) []byte {
	return ctx.Value(contextKey{}).(contextValue).sender
}

// Whoami returns the address of the account being invoked.
func Whoami(ctx context.Context) []byte {
	return ctx.Value(contextKey{}).(contextValue).whoami
}
