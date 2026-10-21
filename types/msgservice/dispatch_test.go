package msgservice_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
	registry := codectypes.NewInterfaceRegistry()
	testdata.RegisterInterfaces(registry)
	router := baseapp.NewMsgServiceRouter()
	router.SetInterfaceRegistry(registry)
	router.SetCircuit(cb)
	testdata.RegisterMsgServer(router, testdata.MsgServerImpl{})
	return router
}

func newCtx() sdk.Context {
	return sdk.Context{}.WithContext(context.Background()).WithEventManager(sdk.NewEventManager())
}

func TestDispatchReturnsTypedResponse(t *testing.T) {
	router := newRouter(t, blockBreaker{})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(), router, msg)
	require.NoError(t, err)
	require.Equal(t, "Spot", resp.Name)
}

func TestDispatchBlockedByCircuitBreaker(t *testing.T) {
	router := newRouter(t, blockBreaker{createDogURL: true})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.MsgCreateDogResponse](newCtx(), router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrUnauthorized)
	require.Nil(t, resp)

	// a direct msg server call skips the router and its breaker
	direct, err := testdata.MsgServerImpl{}.CreateDog(newCtx(), msg)
	require.NoError(t, err)
	require.Equal(t, "Spot", direct.Name)
}

func TestDispatchUnknownMsg(t *testing.T) {
	router := newRouter(t, blockBreaker{})

	resp, err := msgservice.Dispatch[*testdata.TestMsg](newCtx(), router, &testdata.TestMsg{})
	require.ErrorIs(t, err, sdkerrors.ErrUnknownRequest)
	require.ErrorContains(t, err, "/testpb.TestMsg")
	require.Nil(t, resp)
}

func TestDispatchWrongResponseType(t *testing.T) {
	router := newRouter(t, blockBreaker{})
	msg := &testdata.MsgCreateDog{Dog: &testdata.Dog{Name: "Spot"}}

	resp, err := msgservice.Dispatch[*testdata.Dog](newCtx(), router, msg)
	require.ErrorIs(t, err, sdkerrors.ErrInvalidType)
	require.Nil(t, resp)
}
