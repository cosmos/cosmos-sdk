package ante_test

import (
	"testing"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	cbtypes "github.com/cosmos/cosmos-sdk/x/circuit/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/ante"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	ctx           sdk.Context
	mockStoreKey  storetypes.StoreKey
	mockMsgURL    string
	mockclientCtx client.Context
	txBuilder     client.TxBuilder
	encCfg        moduletestutil.TestEncodingConfig
}

type MockCircuitBreaker struct {
	isAllowed bool
}

func (m MockCircuitBreaker) IsAllowed(ctx sdk.Context, typeURL string) bool {
	return typeURL == "/cosmos.circuit.v1.MsgAuthorizeCircuitBreaker"
}

func initFixture(t *testing.T) *fixture {
	mockStoreKey := storetypes.NewKVStoreKey("test")
	mockCtx := testutil.DefaultContextWithDB(t, mockStoreKey, storetypes.NewTransientStoreKey("transient_test"))
	ctx := mockCtx.Ctx.WithBlockHeader(cmproto.Header{})
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{})
	mockMsgURL := "test"
	mockclientCtx := client.Context{}.
		WithTxConfig(encCfg.TxConfig).
		WithClient(clitestutil.NewMockCometRPC(abci.ResponseQuery{}))
	txBuilder := mockclientCtx.TxConfig.NewTxBuilder()

	return &fixture{
		ctx:           ctx,
		mockStoreKey:  mockStoreKey,
		mockMsgURL:    mockMsgURL,
		mockclientCtx: mockclientCtx,
		txBuilder:     txBuilder,
	}
}

func TestCircuitBreakerDecorator(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	// Circuit breaker is allowed to pass through all transactions

	circuitBreaker := &MockCircuitBreaker{true}
	// CircuitBreakerDecorator AnteHandler should always return success
	decorator := ante.NewCircuitBreakerDecorator(circuitBreaker)
	// Test case 1: the transaction is allowed
	// Test case 2: the transaction is not allowed

	msg1 := &cbtypes.MsgAuthorizeCircuitBreaker{
		Grantee: "cosmos1fghij",
		Granter: "cosmos1abcde",
	}

	f.txBuilder.SetMsgs(msg1)
	tx := f.txBuilder.GetTx()

	newCtx, err := decorator.AnteHandle(f.ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})
	require.NoError(t, err)
	require.NotNil(t, newCtx)

	// keys and addresses
	_, _, addr1 := testdata.KeyTestPubAddr()

	// msg and signatures
	msg := testdata.NewTestMsg(addr1)

	f.txBuilder.SetMsgs(msg)
	tx = f.txBuilder.GetTx()

	newCtx, err = decorator.AnteHandle(f.ctx, tx, false, func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, err error) {
		return ctx, nil
	})

	require.Equal(t, "tx type not allowed", err.Error())
	require.NotNil(t, newCtx)
}
