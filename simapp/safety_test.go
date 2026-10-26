package simapp

import (
	"math/rand"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	safetytypes "github.com/cosmos/cosmos-sdk/x/safety/types"
)

// TestSafetyWiring checks the SimApp wiring of x/safety: the default genesis,
// the ante decorator rejecting tripped msgs in CheckTx before fees are taken,
// the router rejecting them in FinalizeBlock after the fee and sequence are
// spent, including msgs nested in authz MsgExec, and reset.
func TestSafetyWiring(t *testing.T) {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(pubKey, 1)})

	priv := secp256k1.GenPrivKey()
	addr := sdk.AccAddress(priv.PubKey().Address())
	acc := authtypes.NewBaseAccount(addr, priv.PubKey(), 0, 0)
	funds := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100_000_000)))
	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, banktypes.Balance{Address: addr.String(), Coins: funds})

	_, err = app.Commit()
	require.NoError(t, err)
	params, err := app.SafetyKeeper.Params.Get(app.NewContext(true))
	require.NoError(t, err)
	require.Equal(t, safetytypes.DefaultParams(authtypes.NewModuleAddress("gov").String()), params)

	// the authority is gov; make addr the operator with bank scope directly in
	// the next block's state, which the first deliver below finalizes and
	// commits (the default scope is empty)
	params.Operator = addr.String()
	params.Scope = []string{"/cosmos.bank.*"}
	ctx := app.NewNextBlockContext(cmtproto.Header{Height: app.LastBlockHeight() + 1})
	require.NoError(t, app.SafetyKeeper.Params.Set(ctx, params))

	txConfig := app.TxConfig()
	fee := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	other := sdk.AccAddress("other_______________")
	send := banktypes.NewMsgSend(addr, other, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))
	exec := authz.NewMsgExec(addr, []sdk.Msg{send})

	encodeFee := func(fee sdk.Coins, msgs ...sdk.Msg) []byte {
		t.Helper()
		ctx := app.NewContext(true)
		a := app.AccountKeeper.GetAccount(ctx, addr)
		tx, err := simtestutil.GenSignedMockTx(rand.New(rand.NewSource(1)), txConfig, msgs, fee,
			simtestutil.DefaultGenTxGas, "", []uint64{a.GetAccountNumber()}, []uint64{a.GetSequence()}, priv)
		require.NoError(t, err)
		bz, err := txConfig.TxEncoder()(tx)
		require.NoError(t, err)
		return bz
	}
	encode := func(msgs ...sdk.Msg) []byte { return encodeFee(fee, msgs...) }
	checkTx := func(bz []byte) *abci.ResponseCheckTx {
		t.Helper()
		res, err := app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: abci.CheckTxType_New})
		require.NoError(t, err)
		return res
	}
	deliver := func(bz []byte) *abci.ExecTxResult {
		t.Helper()
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1, Txs: [][]byte{bz}})
		require.NoError(t, err)
		_, err = app.Commit()
		require.NoError(t, err)
		return res.TxResults[0]
	}
	balance := func() sdkmath.Int {
		return app.BankKeeper.GetBalance(app.NewContext(true), addr, sdk.DefaultBondDenom).Amount
	}

	res := deliver(encode(safetytypes.NewMsgTrip(addr.String(), "/cosmos.bank.v1beta1.MsgSend")))
	require.Zero(t, res.Code, res.Log)

	disabled := safetytypes.ErrMsgDisabled
	for name, msg := range map[string]sdk.Msg{"top level": send, "nested in authz": &exec} {
		bz := encode(msg)
		before := balance()
		cres := checkTx(bz)
		require.Equal(t, disabled.ABCICode(), cres.Code, "%s: %s", name, cres.Log)
		require.Equal(t, disabled.Codespace(), cres.Codespace, name)

		require.Equal(t, before, balance(), "%s: no fee may be taken in CheckTx", name)

		// in FinalizeBlock the router rejects it after the fee and sequence
		// are spent, so the bytes cannot run again after a reset
		dres := deliver(bz)
		require.Equal(t, disabled.ABCICode(), dres.Code, "%s: %s", name, dres.Log)
		require.Equal(t, before.Sub(fee.AmountOf(sdk.DefaultBondDenom)), balance(), "%s: the fee must be taken", name)
		require.Equal(t, sdkerrors.ErrWrongSequence.ABCICode(), deliver(bz).Code, name)
	}

	// the decorator runs before fee deduction: a tripped tx whose fee the
	// sender cannot pay fails with ErrMsgDisabled, not insufficient funds
	unpayable := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, balance().AddRaw(1)))
	cres := checkTx(encodeFee(unpayable, send))
	require.Equal(t, disabled.ABCICode(), cres.Code, cres.Log)
	require.Equal(t, disabled.Codespace(), cres.Codespace, cres.Log)

	// a protected msg still passes: the operator resets
	res = deliver(encode(safetytypes.NewMsgReset(addr.String(), "/cosmos.bank.v1beta1.MsgSend")))
	require.Zero(t, res.Code, res.Log)

	bz := encode(send)
	require.Zero(t, checkTx(bz).Code)
	res = deliver(bz)
	require.Zero(t, res.Code, res.Log)
	require.Equal(t, int64(1), app.BankKeeper.GetBalance(app.NewContext(true), other, sdk.DefaultBondDenom).Amount.Int64())
}

// TestSafetyGatesHybridHandlers checks that SimApp's hybrid handlers run the
// x/safety check for an exact, a prefix and a "*" trip. HybridHandlerByMsgName
// is deprecated, and callers keep it until a future release removes it.
func TestSafetyGatesHybridHandlers(t *testing.T) {
	app := Setup(t, false)
	ctx := app.NewContext(false)
	from := sdk.AccAddress("from________________")
	to := sdk.AccAddress("to__________________")
	send := banktypes.NewMsgSend(from, to, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)))

	handler := app.MsgServiceRouter().HybridHandlerByMsgName(sdk.MsgTypeURL(send)[1:]) //nolint:staticcheck // covers the deprecated path while it ships
	require.NotNil(t, handler)

	for _, p := range []string{sdk.MsgTypeURL(send), "/cosmos.bank.*", safetytypes.Wildcard} {
		require.NoError(t, app.SafetyKeeper.DisabledList.Set(ctx, p))
		err := handler(ctx, send, &banktypes.MsgSendResponse{})
		require.ErrorIs(t, err, safetytypes.ErrMsgDisabled, p)
		require.NoError(t, app.SafetyKeeper.DisabledList.Remove(ctx, p))
	}

	// control: with nothing disabled the bank handler runs, and fails on the
	// unfunded sender
	err := handler(ctx, send, &banktypes.MsgSendResponse{})
	require.ErrorIs(t, err, sdkerrors.ErrInsufficientFunds)
}

// TestRecheckEvictsPreTripTx shows what happens to a tx that entered the
// mempool before a trip. ReCheckTx runs the decorator, so it turns the tx
// away before it pays a fee or uses its sequence, and its signed bytes
// execute once the trip is reset.
func TestRecheckEvictsPreTripTx(t *testing.T) {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)
	valSet := cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(pubKey, 1)})

	opPriv, userPriv := secp256k1.GenPrivKey(), secp256k1.GenPrivKey()
	op, user := sdk.AccAddress(opPriv.PubKey().Address()), sdk.AccAddress(userPriv.PubKey().Address())
	funds := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100_000_000)))
	app := SetupWithGenesisValSet(t, valSet,
		[]authtypes.GenesisAccount{authtypes.NewBaseAccount(op, opPriv.PubKey(), 0, 0), authtypes.NewBaseAccount(user, userPriv.PubKey(), 1, 0)},
		banktypes.Balance{Address: op.String(), Coins: funds}, banktypes.Balance{Address: user.String(), Coins: funds})
	_, err = app.Commit()
	require.NoError(t, err)

	params, err := app.SafetyKeeper.Params.Get(app.NewContext(true))
	require.NoError(t, err)
	params.Operator = op.String()
	params.Scope = []string{"/cosmos.bank.*"}
	require.NoError(t, app.SafetyKeeper.Params.Set(app.NewNextBlockContext(cmtproto.Header{Height: app.LastBlockHeight() + 1}), params))

	txConfig := app.TxConfig()
	fee := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	encode := func(priv *secp256k1.PrivKey, msg sdk.Msg) []byte {
		t.Helper()
		a := app.AccountKeeper.GetAccount(app.NewContext(true), sdk.AccAddress(priv.PubKey().Address()))
		tx, err := simtestutil.GenSignedMockTx(rand.New(rand.NewSource(1)), txConfig, []sdk.Msg{msg}, fee,
			simtestutil.DefaultGenTxGas, "", []uint64{a.GetAccountNumber()}, []uint64{a.GetSequence()}, priv)
		require.NoError(t, err)
		bz, err := txConfig.TxEncoder()(tx)
		require.NoError(t, err)
		return bz
	}
	checkTx := func(bz []byte, typ abci.CheckTxType) *abci.ResponseCheckTx {
		t.Helper()
		res, err := app.CheckTx(&abci.RequestCheckTx{Tx: bz, Type: typ})
		require.NoError(t, err)
		return res
	}
	deliver := func(bz []byte) *abci.ExecTxResult {
		t.Helper()
		res, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1, Txs: [][]byte{bz}})
		require.NoError(t, err)
		_, err = app.Commit()
		require.NoError(t, err)
		return res.TxResults[0]
	}
	sequence := func() uint64 {
		return app.AccountKeeper.GetAccount(app.NewContext(true), user).GetSequence()
	}
	balance := func() sdkmath.Int {
		return app.BankKeeper.GetBalance(app.NewContext(true), user, sdk.DefaultBondDenom).Amount
	}

	const sendURL = "/cosmos.bank.v1beta1.MsgSend"
	other := sdk.AccAddress("other_______________")
	sendBz := encode(userPriv, banktypes.NewMsgSend(user, other, sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))))
	require.Zero(t, checkTx(sendBz, abci.CheckTxType_New).Code)

	res := deliver(encode(opPriv, safetytypes.NewMsgTrip(op.String(), sendURL)))
	require.Zero(t, res.Code, res.Log)

	before := balance()
	rres := checkTx(sendBz, abci.CheckTxType_Recheck)
	require.Equal(t, safetytypes.ErrMsgDisabled.ABCICode(), rres.Code, rres.Log)
	require.Equal(t, safetytypes.ErrMsgDisabled.Codespace(), rres.Codespace, rres.Log)
	require.Zero(t, sequence(), "the evicted tx must not use its sequence")
	require.Equal(t, before, balance(), "the evicted tx must not pay its fee")

	res = deliver(encode(opPriv, safetytypes.NewMsgReset(op.String(), sendURL)))
	require.Zero(t, res.Code, res.Log)

	res = deliver(sendBz)
	require.Zero(t, res.Code, "the pre-trip bytes execute after the reset: %s", res.Log)
	require.Equal(t, uint64(1), sequence())
}
