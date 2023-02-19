package ante_test

import (
	storetypes "cosmossdk.io/store/types"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/ante"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	ctx          sdk.Context
	mockStoreKey storetypes.StoreKey
}

type MockCircuitBreaker struct {
	isAllowed bool
}

func (m MockCircuitBreaker) IsAllowed(ctx sdk.Context, typeURL string) bool {
	// Here you can implement the condition to return `true` or `false`
	return m.isAllowed
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

func TestCircuitBreakerDecorator(t *testing.T) {
	f := initFixture(t)

	// Circuit breaker is allowed to pass through all transactions
	circuitBreaker := &MockCircuitBreaker{true}
	// CircuitBreakerDecorator AnteHandler should always return success
	decorator := ante.NewCircuitBreakerDecorator(circuitBreaker)

	// Test case 1: the transaction is allowed
	newCtx, err := decorator.AnteHandle(f.ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})
	require.NoError(t, err)
	require.NotNil(t, newCtx)

	// Test case 2: the transaction is not allowed
	circuitBreaker.isAllowed = false
	decorator = ante.NewCircuitBreakerDecorator(circuitBreaker)

	newCtx, err = decorator.AnteHandle(f.ctx, nil, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})

	require.Error(t, err)
	require.Equal(t, "tx type not allowed", err.Error())
	require.NotNil(t, newCtx)
}
