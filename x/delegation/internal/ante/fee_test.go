package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/delegation/internal/ante"
	// delTypes "github.com/cosmos/cosmos-sdk/x/delegation/internal/types"
)

func TestDeductFeesNoDelegation(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := authtypes.KeyTestPubAddr()

	// msg and signatures
	msg1 := authtypes.NewTestMsg(addr1)
	fee := authtypes.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := authtypes.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	// Set account with insufficient funds
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc.SetCoins([]sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(10))})
	app.AccountKeeper.SetAccount(ctx, acc)

	dfd := ante.NewDeductDelegatedFeeDecorator(app.AccountKeeper, app.SupplyKeeper, app.DelegationKeeper)
	antehandler := sdk.ChainAnteDecorators(dfd)

	_, err := antehandler(ctx, tx, false)

	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
	acc.SetCoins([]sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(200))})
	app.AccountKeeper.SetAccount(ctx, acc)

	_, err = antehandler(ctx, tx, false)

	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")
}

// returns context and app with params set on account keeper
func createTestApp(isCheckTx bool) (*simapp.SimApp, sdk.Context) {
	app := simapp.Setup(isCheckTx)
	ctx := app.BaseApp.NewContext(isCheckTx, abci.Header{})
	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return app, ctx
}
