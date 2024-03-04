package implementation

import (
	"context"
	"encoding/binary"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/prefixstore"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var AccountStatePrefix = collections.NewPrefix(255)

type (
	ModuleExecUntypedFunc = func(ctx context.Context, sender []byte, msg ProtoMsg) (ProtoMsg, error)
	ModuleExecFunc        = func(ctx context.Context, sender []byte, msg, msgResp ProtoMsg) error
	ModuleQueryFunc       = func(ctx context.Context, queryReq, queryResp ProtoMsg) error
)

type contextKey struct{}

type contextValue struct {
	store             store.KVStore         // store is the prefixed store for the account.
	sender            []byte                // sender is the address of the entity invoking the account action.
	whoami            []byte                // whoami is the address of the account being invoked.
	funds             sdk.Coins             // funds reports the coins sent alongside the request.
	parentContext     context.Context       // parentContext that was used to build the account context.
	moduleExec        ModuleExecFunc        // moduleExec is a function that executes a module message, when the resp type is known.
	moduleExecUntyped ModuleExecUntypedFunc // moduleExecUntyped is a function that executes a module message, when the resp type is unknown.
	moduleQuery       ModuleQueryFunc       // moduleQuery is a function that queries a module.
}

func addCtx(ctx context.Context, value contextValue) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

func getCtx(ctx context.Context) contextValue {
	return ctx.Value(contextKey{}).(contextValue)
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
	accNumber uint64,
	accountAddr []byte,
	sender []byte,
	funds sdk.Coins,
	moduleExec ModuleExecFunc,
	moduleExecUntyped ModuleExecUntypedFunc,
	moduleQuery ModuleQueryFunc,
) context.Context {
	return addCtx(ctx, contextValue{
		store:             makeAccountStore(ctx, storeSvc, accNumber),
		sender:            sender,
		whoami:            accountAddr,
		funds:             funds,
		parentContext:     ctx,
		moduleExec:        moduleExec,
		moduleExecUntyped: moduleExecUntyped,
		moduleQuery:       moduleQuery,
	})
}

// makeAccountStore creates the prefixed store for the account.
// It uses the number of the account, this gives constant size
// bytes prefixes for the account state.
func makeAccountStore(ctx context.Context, storeSvc store.KVStoreService, accNum uint64) store.KVStore {
	prefix := make([]byte, 8)
	binary.BigEndian.PutUint64(prefix, accNum)
	return prefixstore.New(storeSvc.OpenKVStore(ctx), append(AccountStatePrefix, prefix...))
}

// ExecModuleUntyped can be used to execute a message towards a module, when the response type is unknown.
func ExecModuleUntyped(ctx context.Context, msg ProtoMsg) (ProtoMsg, error) {
	// get sender
	v := getCtx(ctx)

	resp, err := v.moduleExecUntyped(v.parentContext, v.whoami, msg)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ExecModule can be used to execute a message towards a module.
func ExecModule[Resp any, RespProto ProtoMsgG[Resp], Req any, ReqProto ProtoMsgG[Req]](ctx context.Context, msg ReqProto) (RespProto, error) {
	// get sender
	v := getCtx(ctx)

	// execute module, unwrapping the original context.
	resp := RespProto(new(Resp))
	err := v.moduleExec(v.parentContext, v.whoami, msg, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// QueryModule can be used by an account to execute a module query.
func QueryModule[Resp any, RespProto ProtoMsgG[Resp], Req any, ReqProto ProtoMsgG[Req]](ctx context.Context, req ReqProto) (RespProto, error) {
	// we do not need to check the sender in a query because it is not a state transition.
	// we also unwrap the original context.
	v := getCtx(ctx)
	resp := RespProto(new(Resp))
	err := v.moduleQuery(v.parentContext, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// openKVStore returns the prefixed store for the account given the context.
func openKVStore(ctx context.Context) store.KVStore { return getCtx(ctx).store }

// Sender returns the address of the entity invoking the account action.
func Sender(ctx context.Context) []byte {
	return getCtx(ctx).sender
}

// Whoami returns the address of the account being invoked.
func Whoami(ctx context.Context) []byte {
	return getCtx(ctx).whoami
}

// Funds returns the funds associated with the execution context.
func Funds(ctx context.Context) sdk.Coins { return getCtx(ctx).funds }
