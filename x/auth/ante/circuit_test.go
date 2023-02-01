package ante

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

type fixture struct {
	ctx          sdk.Context
	mockStoreKey storetypes.StoreKey
}

func initFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	mockCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := mockCtx.Ctx.WithBlockHeader(tmproto.Header{})

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

	// Test case 1: the circuit breaker is open
	_, err := decorator.AnteHandle(f.ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})

	require.Equal(t, sdkerrors.Wrap(sdkerrors.ErrActivation, "transaction is blocked due to Circuit Breaker activation"), err)

	// Test case 2: the circuit breaker is closed
	mockKeeper.circuitOpen = false
	_, err = decorator.AnteHandle(f.ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})
	require.NoError(t, err)
}
