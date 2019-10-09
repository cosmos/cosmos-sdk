package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestEnsureMempoolFees(t *testing.T) {
	// setup
	_, ctx := createTestApp(true)

	mfd := ante.NewMempoolFeeDecorator()
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	// Set high gas price so standard test fee fails
	atomPrice := sdk.NewDecCoinFromDec("atom", sdk.NewDec(200).Quo(sdk.NewDec(100000)))
	highGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(highGasPrice)

	// Set IsCheckTx to true
	ctx = ctx.WithIsCheckTx(true)

	// antehandler errors with insufficient fees
	_, err := antehandler(ctx, tx, false)
	require.NotNil(t, err, "Decorator should have errored on too low fee for local gasPrice")

	// Set IsCheckTx to false
	ctx = ctx.WithIsCheckTx(false)

	// antehandler should not error since we do not check minGasPrice in DeliverTx
	_, err = antehandler(ctx, tx, false)
	require.Nil(t, err, "MempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	ctx = ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(ctx, tx, false)
	require.Nil(t, err, "Decorator should not have errored on fee higher than local gasPrice")
}

func TestDeductFees(t *testing.T) {
	// setup
	app, ctx := createTestApp(true)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// msg and signatures
	msg1 := types.NewTestMsg(addr1)
	fee := types.NewTestStdFee()

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	tx := types.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

	// Set account with insufficient funds
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	acc.SetCoins([]sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(10))})
	app.AccountKeeper.SetAccount(ctx, acc)

	dfd := ante.NewDeductFeeDecorator(app.AccountKeeper, app.SupplyKeeper)
	antehandler := sdk.ChainAnteDecorators(dfd)

	_, err := antehandler(ctx, tx, false)

	require.NotNil(t, err, "Tx did not error when fee payer had insufficient funds")

	// Set account with sufficient funds
	acc.SetCoins([]sdk.Coin{sdk.NewCoin("atom", sdk.NewInt(200))})
	app.AccountKeeper.SetAccount(ctx, acc)

	_, err = antehandler(ctx, tx, false)

	require.Nil(t, err, "Tx errored after account has been set with sufficient funds")
}
