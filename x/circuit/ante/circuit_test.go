package ante_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth"
	"cosmossdk.io/x/circuit/ante"
	cbtypes "cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

type fixture struct {
	ctx           context.Context
	mockStoreKey  storetypes.StoreKey
	mockMsgURL    string
	mockclientCtx client.Context
	txBuilder     client.TxBuilder
}

type MockCircuitBreaker struct {
	isAllowed bool
}

func (m MockCircuitBreaker) IsAllowed(ctx context.Context, typeURL string) (bool, error) {
	return typeURL == "/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker", nil
}

func initFixture(t *testing.T) *fixture {
	t.Helper()
	mockStoreKey := storetypes.NewKVStoreKey("test")
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModule{})
	mockclientCtx := client.Context{}.
		WithTxConfig(encCfg.TxConfig)

	return &fixture{
		ctx:           testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx,
		mockStoreKey:  mockStoreKey,
		mockMsgURL:    "test",
		mockclientCtx: mockclientCtx,
		txBuilder:     mockclientCtx.TxConfig.NewTxBuilder(),
	}
}

func TestCircuitBreakerDecorator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	_, _, addr1 := testdata.KeyTestPubAddr()

	testcases := []struct {
		msg     sdk.Msg
		allowed bool
	}{
		{msg: &cbtypes.MsgAuthorizeCircuitBreaker{
			Grantee: "cosmos139f7kncmglres2nf3h4hc4tade85ekfr8sulz5",
			Granter: "cosmos16wfryel63g7axeamw68630wglalcnk3l0zuadc",
		}, allowed: true},
		{msg: testdata.NewTestMsg(addr1), allowed: false},
	}

	for _, tc := range testcases {
		// Circuit breaker is allowed to pass through all transactions
		circuitBreaker := &MockCircuitBreaker{true}
		// CircuitBreakerDecorator AnteHandler should always return success
		decorator := ante.NewCircuitBreakerDecorator(circuitBreaker)

		err := f.txBuilder.SetMsgs(tc.msg)
		require.NoError(t, err)

		tx := f.txBuilder.GetTx()

		sdkCtx := sdk.UnwrapSDKContext(f.ctx)
		_, err = decorator.AnteHandle(sdkCtx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			return ctx, nil
		})

		if tc.allowed {
			require.NoError(t, err)
		} else {
			require.Equal(t, "tx type not allowed", err.Error())
		}
	}
}
