package distribution

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/multisig"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)


// Test logic around fee deduction.
func TestAnteHandlerFees(t *testing.T) {
	// setup
	input := setupTestInput()
	ctx := input.ctx
	anteHandler := NewAnteHandler(input.sk)

	// keys and addresses
	priv1, _, addr1 := types.KeyTestPubAddr()

	// set the accounts
	acc1 := input.ak.NewAccountWithAddress(ctx, addr1)
	input.ak.SetAccount(ctx, acc1)

	// msg and signatures
	var tx sdk.Tx
	msg := types.NewTestMsg(addr1)
	privs, accnums, seqs := []crypto.PrivKey{priv1}, []uint64{0}, []uint64{0}
	fee := types.NewTestStdFee()
	msgs := []sdk.Msg{msg}

	// signer does not have enough funds to pay the fee
	tx = types.NewTestTx(ctx, msgs, privs, accnums, seqs, fee)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInsufficientFunds)

	acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 149)))
	input.ak.SetAccount(ctx, acc1)
	checkInvalidTx(t, anteHandler, ctx, tx, false, sdk.CodeInsufficientFunds)

	require.True(t.GetModuleAccount(ctx, types.FeeCollectorName).GetCoins().Empty())
	require.True(sdk.IntEq(t, input.ak.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), sdk.NewInt(149)))

	acc1.SetCoins(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	input.ak.SetAccount(ctx, acc1)
	checkValidTx(t, anteHandler, ctx, tx, false)

	require.True(sdk.IntEq(t.GetModuleAccount(ctx, types.FeeCollectorName).GetCoins().AmountOf("atom"), sdk.NewInt(150)))
	require.True(sdk.IntEq(t, input.ak.GetAccount(ctx, addr1).GetCoins().AmountOf("atom"), sdk.NewInt(0)))
}


func TestEnsureSufficientMempoolFees(t *testing.T) {
	// setup
	input := setupTestInput()
	ctx := input.ctx.WithMinGasPrices(
		sdk.DecCoins{
			sdk.NewDecCoinFromDec("photino", sdk.NewDecWithPrec(50000000000000, sdk.Precision)), // 0.0001photino
			sdk.NewDecCoinFromDec("stake", sdk.NewDecWithPrec(10000000000000, sdk.Precision)),   // 0.000001stake
		},
	)

	testCases := []struct {
		input      StdFee
		expectedOK bool
	}{
		{NewStdFee(200000, sdk.Coins{}), false},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("photino", 5))), false},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("stake", 1))), false},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("stake", 2))), true},
		{NewStdFee(200000, sdk.NewCoins(sdk.NewInt64Coin("photino", 10))), true},
		{
			NewStdFee(
				200000,
				sdk.NewCoins(
					sdk.NewInt64Coin("photino", 10),
					sdk.NewInt64Coin("stake", 2),
				),
			),
			true,
		},
		{
			NewStdFee(
				200000,
				sdk.NewCoins(
					sdk.NewInt64Coin("atom", 5),
					sdk.NewInt64Coin("photino", 10),
					sdk.NewInt64Coin("stake", 2),
				),
			),
			true,
		},
	}

	for i, tc := range testCases {
		res := EnsureSufficientMempoolFees(ctx, tc.input)
		require.Equal(
			t, tc.expectedOK, res.IsOK(),
			"unexpected result; tc #%d, input: %v, log: %v", i, tc.input, res.Log,
		)
	}
}
