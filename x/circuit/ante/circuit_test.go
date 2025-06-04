package ante_test

import (
	"context"
	"testing"

	cmproto "github.com/cometbft/cometbft/api/cometbft/types/v2"
	abci "github.com/cometbft/cometbft/v2/abci/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/circuit/ante"
	cbtypes "github.com/cosmos/cosmos-sdk/x/circuit/types"
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
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	mockclientCtx := client.Context{}.
		WithTxConfig(encCfg.TxConfig).
		WithClient(clitestutil.NewMockCometRPC(abci.QueryResponse{}))

	return &fixture{
		ctx:           testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test")).Ctx.WithBlockHeader(cmproto.Header{}),
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
			Grantee: "cosmos1fghij",
			Granter: "cosmos1abcde",
		}, allowed: true},
		{msg: testdata.NewTestMsg(addr1), allowed: false},
	}

	for _, tc := range testcases {
		// Circuit breaker is allowed to pass through all transactions
		circuitBreaker := &MockCircuitBreaker{true}
		// CircuitBreakerDecorator AnteHandler should always return success
		decorator := ante.NewCircuitBreakerDecorator(circuitBreaker)

		require.NoError(t, f.txBuilder.SetMsgs(tc.msg))
		tx := f.txBuilder.GetTx()

		sdkCtx := sdk.UnwrapSDKContext(f.ctx)
		_, err := decorator.AnteHandle(sdkCtx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
			return ctx, nil
		})

		if tc.allowed {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
			require.Equal(t, "tx type not allowed", err.Error())
		}
	}
}
