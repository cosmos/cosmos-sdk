package implementation

import (
	"context"

	"google.golang.org/protobuf/proto"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/prefixstore"
)

var AccountStatePrefix = collections.NewPrefix(255)

type (
	ModuleExecUntypedFunc = func(ctx context.Context, sender []byte, msg proto.Message) (proto.Message, error)
	ModuleExecFunc        = func(ctx context.Context, sender []byte, msg, msgResp proto.Message) error
	ModuleQueryFunc       = func(ctx context.Context, queryReq, queryResp proto.Message) error
)

type contextKey struct{}

type contextValue struct {
	store             store.KVStore         // store is the prefixed store for the account.
	sender            []byte                // sender is the address of the entity invoking the account action.
	whoami            []byte                // whoami is the address of the account being invoked.
	originalContext   context.Context       // originalContext that was used to build the account context.
	moduleExec        ModuleExecFunc        // moduleExec is a function that executes a module message, when the resp type is known.
	moduleExecUntyped ModuleExecUntypedFunc // moduleExecUntyped is a function that executes a module message, when the resp type is unknown.
	moduleQuery       ModuleQueryFunc       // moduleQuery is a function that queries a module.
}

// MakeAccountContext creates a new account execution context given:
// storeSvc: which fetches the x/accounts module store.
// accountAddr: the address of the account being invoked, which is used to give the
// account a prefixed storage.
// sender: the address of entity invoking the account action.
// moduleExec: a function that executes a module message.
// moduleQuery: a function that queries a module.
func MakeAccountContext(
	ctx context.Context,
	storeSvc store.KVStoreService,
	accountAddr,
	sender []byte,
	moduleExec ModuleExecFunc,
	moduleExecUntyped ModuleExecUntypedFunc,
	moduleQuery ModuleQueryFunc,
) context.Context {
	return context.WithValue(ctx, contextKey{}, contextValue{
		store:             prefixstore.New(storeSvc.OpenKVStore(ctx), append(AccountStatePrefix, accountAddr...)),
		sender:            sender,
		whoami:            accountAddr,
		originalContext:   ctx,
		moduleExecUntyped: moduleExecUntyped,
		moduleExec:        moduleExec,
		moduleQuery:       moduleQuery,
	})
}

// ExecModuleUntyped can be used to execute a message towards a module, when the response type is unknown.
func ExecModuleUntyped(ctx context.Context, msg proto.Message) (proto.Message, error) {
	// get sender
	v := ctx.Value(contextKey{}).(contextValue)

	resp, err := v.moduleExecUntyped(v.originalContext, v.whoami, msg)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExecModule can be used to execute a message towards a module.
func ExecModule[Resp any, RespProto ProtoMsg[Resp], Req any, ReqProto ProtoMsg[Req]](ctx context.Context, msg ReqProto) (RespProto, error) {
	// get sender
	v := ctx.Value(contextKey{}).(contextValue)

	// execute module, unwrapping the original context.
	resp := RespProto(new(Resp))
	err := v.moduleExec(v.originalContext, v.whoami, msg, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// QueryModule can be used by an account to execute a module query.
func QueryModule[Resp any, RespProto ProtoMsg[Resp], Req any, ReqProto ProtoMsg[Req]](ctx context.Context, req ReqProto) (RespProto, error) {
	// we do not need to check the sender in a query because it is not a state transition.
	// we also unwrap the original context.
	v := ctx.Value(contextKey{}).(contextValue)
	resp := RespProto(new(Resp))
	err := v.moduleQuery(v.originalContext, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
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
