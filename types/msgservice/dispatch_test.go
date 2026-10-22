package msgservice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

const createDogURL = "/testpb.MsgCreateDog"

// blockBreaker blocks the type URLs in blocked.
type blockBreaker map[string]bool

func (b blockBreaker) IsAllowed(_ context.Context, typeURL string) (bool, error) {
	if b[typeURL] {
		return false, sdkerrors.ErrUnauthorized.Wrap(typeURL)
	}
	return true, nil
}

func newRouter(t *testing.T, cb sdk.CircuitBreaker) *baseapp.MsgServiceRouter {
	t.Helper()
	return newRouterWithServer(t, cb, testdata.MsgServerImpl{})
}

func newRouterWithServer(t *testing.T, cb sdk.CircuitBreaker, srv testdata.MsgServer) *baseapp.MsgServiceRouter {
	t.Helper()
	registry := codectypes.NewInterfaceRegistry()
	testdata.RegisterInterfaces(registry)
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(registry)
	router.SetCircuit(cb)
	testdata.RegisterMsgServer(router, srv)
	return router
}

func newCtx(t *testing.T) sdk.Context {
	t.Helper()
	ctx, _ := newStoreCtx(t)
	return ctx
}

func TestDispatchReturnsTypedResponse(t *testing.T) {
	router := newRouter(t, blockBreaker{})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(t), router, msg)
	require.NoError(t, err)
	require.Equal(t, "Spot", resp.Name)
}

func TestDispatchBlockedByCircuitBreaker(t *testing.T) {
	router := newRouter(t, blockBreaker{createDogURL: true})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(t), router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.Nil(t, resp)
}

func TestDispatchUnknownMsg(t *testing.T) {
	router := newRouter(t, blockBreaker{})

	resp, err := msgservice.Dispatch[*testdata.TestMsg](newCtx(t), router, &testdata.TestMsg{})
	require.ErrorIs(t, err, sdkerrors.ErrUnknownRequest)
	require.ErrorContains(t, err, "/testpb.TestMsg")
	require.Nil(t, resp)
}

// TestDispatchPulsarMsg dispatches the pulsar type of a msg. The router
// handler for its type URL takes the gogo type, so Dispatch must return an
// error instead of letting the handler panic.
func TestDispatchPulsarMsg(t *testing.T) {
	router := newRouter(t, blockBreaker{})
	msg := &testpb.MsgCreateDog{Dog: &testpb.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(t), router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidType)
	require.ErrorContains(t, err, createDogURL)
	require.Nil(t, resp)
}

func TestDispatchWrongResponseType(t *testing.T) {
	router := newRouter(t, blockBreaker{})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.Dog](newCtx(t), router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidType)
	require.Nil(t, resp)
}

// writeServer writes to its store and emits an event in CreateDog, then
// returns err.
type writeServer struct {
	key storetypes.StoreKey
	err error
}

func (s writeServer) CreateDog(ctx context.Context, msg *testdata.MsgCreateDog) (*testdata.MsgCreateDogResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.KVStore(s.key).Set([]byte("k"), []byte("v"))
	sdkCtx.EventManager().EmitEvent(sdk.NewEvent("dog"))
	return &testdata.MsgCreateDogResponse{Name: msg.Dog.Name}, s.err
}

func newStoreCtx(t *testing.T) (sdk.Context, storetypes.StoreKey) {
	t.Helper()
	key := storetypes.NewKVStoreKey("dispatch")
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_dispatch")).Ctx
	return ctx.WithEventManager(sdk.NewEventManager()), key
}

func TestDispatchWritesOnSuccess(t *testing.T) {
	ctx, key := newStoreCtx(t)
	router := newRouterWithServer(t, blockBreaker{}, writeServer{key: key})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](ctx, router, msg)
	require.NoError(t, err)
	require.Equal(t, "Spot", resp.Name)
	require.Equal(t, []byte("v"), ctx.KVStore(key).Get([]byte("k")))
	require.Len(t, ctx.EventManager().Events(), 1)
}

func TestDispatchHandlerErrorLeavesCtxUnchanged(t *testing.T) {
	ctx, key := newStoreCtx(t)
	router := newRouterWithServer(t, blockBreaker{}, writeServer{key: key, err: sdkerrors.ErrInvalidRequest.Wrap("fail after write")})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	_, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](ctx, router, msg)
	require.ErrorContains(t, err, "fail after write")
	require.Nil(t, ctx.KVStore(key).Get([]byte("k")))
	require.Empty(t, ctx.EventManager().Events())
}

func TestDispatchWrongResponseTypeLeavesCtxUnchanged(t *testing.T) {
	ctx, key := newStoreCtx(t)
	router := newRouterWithServer(t, blockBreaker{}, writeServer{key: key})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.Dog](ctx, router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidType)
	require.Nil(t, resp)
	require.Nil(t, ctx.KVStore(key).Get([]byte("k")))
	require.Empty(t, ctx.EventManager().Events())
}

func TestDispatchNilRouter(t *testing.T) {
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(t), nil, msg)
	require.ErrorIs(t, err, sdkerrors.ErrLogic)
	require.Nil(t, resp)

	var router *baseapp.MsgServiceRouter
	resp, err = msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(t), router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrUnknownRequest)
	require.Nil(t, resp)
}
