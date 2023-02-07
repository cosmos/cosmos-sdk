package ante

import (
	storetypes "cosmossdk.io/store/types"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

type fixture struct {
	ctx          sdk.Context
	mockStoreKey storetypes.StoreKey
}

func initFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	mockCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := mockCtx.Ctx.WithBlockHeader(cmproto.Header{})

	return &fixture{
		ctx:          ctx,
		mockStoreKey: mockStoreKey,
	}
}

type mockCircuitBreakerKeeper struct {
	circuitOpen bool
}

func (k *mockCircuitBreakerKeeper) IsCircuitOpen(ctx sdk.Context) bool {
	return k.circuitOpen
}

func (k *mockCircuitBreakerKeeper) OpenCircuit(ctx sdk.Context) {
	k.circuitOpen = true
}

func (k *mockCircuitBreakerKeeper) CloseCircuit(ctx sdk.Context) {
	k.circuitOpen = false
}

func TestCircuitBreakerDecorator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	mockKeeper := &mockCircuitBreakerKeeper{circuitOpen: true}
	decorator := NewCircuitBreakerDecorator(mockKeeper)
	expectedError := sdkerrors.Wrap(sdkerrors.ErrActivation, "transaction is blocked due to Circuit Breaker activation")

	// Test case 1: the circuit breaker is open
	_, err := decorator.AnteHandle(f.ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})

	require.Equal(t, expectedError.Error(), err.Error())

	// Test case 2: the circuit breaker is closed
	mockKeeper.circuitOpen = false
	_, err = decorator.AnteHandle(f.ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})
	require.NoError(t, err)
}

//
//func TestCircuitBreaker(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	defer ctrl.Finish()
//
//	keeper := testutil.NewMockCircuitBreakerKeeper(ctrl)
//
//	// Use the expect method to set up your expectations for how the mock
//	// should behave during the test.
//	keeper.EXPECT().IsCircuitOpen(gomock.Any()).Return(false)
//	keeper.EXPECT().OpenCircuit(gomock.Any()).Times(1)
//	keeper.EXPECT().CloseCircuit(gomock.Any()).Times(1)
//
//	// Call the function that you want to test.
//	TestCircuitBreakerDecorator(keeper)
//
//	// Use the require package to make assertions about the behavior of the code.
//	require.False(t, keeper.IsCircuitOpen(gomock.Any()))
//}
//
//func TestCircuitBreakerDecorator(t *testing.T) {
//	ctx := types.NewContext(nil, types.Header{}, false, nil)
//	decorator := NewCircuitBreakerDecorator(&mockCircuitBreakerKeeper{isOpen: false})
//
//	// Test that the function proceeds as normal when the circuit is closed
//	newCtx, err := decorator.AnteHandle(ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
//		return ctx, nil
//	})
//	require.NoError(t, err)
//	require.Equal(t, ctx, newCtx)
//
//	// Test that the function returns an error when the circuit is open
//	newCtx, err = decorator.AnteHandle(ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
//		return ctx, nil
//	})
//	require.Error(t, err)
//	require.Equal(t, sdkerrors.Code(err), sdkerrors.Code(sdkerrors.ErrActivation))
//	require.Equal(t, ctx, newCtx)
//}
