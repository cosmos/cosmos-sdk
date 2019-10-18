package ante_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	// "github.com/cosmos/cosmos-sdk/x/subkeys/internal/types"
	"github.com/cosmos/cosmos-sdk/x/subkeys/internal/ante"
)

func TestEnsureMempoolFees(t *testing.T) {
	// setup
	_, ctx := createTestApp(true)

	mfd := ante.NewDelegatedMempoolFeeDecorator()
	antehandler := sdk.ChainAnteDecorators(mfd)

	// keys and addresses
	priv1, _, addr1 := authtypes.KeyTestPubAddr()

	// msg and signatures
	msg1 := authtypes.NewTestMsg(addr1)
	fee := authtypes.NewStdFee(100000, sdk.NewCoins(sdk.NewInt64Coin("atom", 100)))

	msgs := []sdk.Msg{msg1}

	privs, accNums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	// TODO
	tx := authtypes.NewTestTx(ctx, msgs, privs, accNums, seqs, fee)

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
	require.Nil(t, err, "DelegatedMempoolFeeDecorator returned error in DeliverTx")

	// Set IsCheckTx back to true for testing sufficient mempool fee
	ctx = ctx.WithIsCheckTx(true)

	atomPrice = sdk.NewDecCoinFromDec("atom", sdk.NewDec(0).Quo(sdk.NewDec(100000)))
	lowGasPrice := []sdk.DecCoin{atomPrice}
	ctx = ctx.WithMinGasPrices(lowGasPrice)

	_, err = antehandler(ctx, tx, false)
	require.Nil(t, err, "Decorator should not have errored on fee higher than local gasPrice")
}
