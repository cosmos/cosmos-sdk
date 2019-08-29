package simulation

import (
	"errors"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulateMsgSend tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSend(ak types.AccountKeeper, bk keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		fromAcc, comment, msg, ok := createMsgSend(r, ctx, accs, ak)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(comment)
		}

		err = sendAndVerifyMsgSend(r, app, ak, msg, ctx, chainID, []crypto.PrivKey{fromAcc.PrivKey})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, comment), nil, nil
	}
}

func createMsgSend(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, ak types.AccountKeeper) (
	fromAcc simulation.Account, comment string, msg types.MsgSend, ok bool) {

	fromAcc = simulation.RandomAcc(r, accs)
	toAcc := simulation.RandomAcc(r, accs)
	// Disallow sending money to yourself
	for {
		if !fromAcc.PubKey.Equals(toAcc.PubKey) {
			break
		}
		toAcc = simulation.RandomAcc(r, accs)
	}
	initFromCoins := ak.GetAccount(ctx, fromAcc.Address).SpendableCoins(ctx.BlockHeader().Time)

	if len(initFromCoins) == 0 {
		return fromAcc, "skipping, no coins at all", msg, false
	}

	denomIndex := r.Intn(len(initFromCoins))
	amt, err := simulation.RandPositiveInt(r, initFromCoins[denomIndex].Amount)
	if err != nil {
		return fromAcc, "skipping bank send due to account having no coins of denomination " + initFromCoins[denomIndex].Denom, msg, false
	}

	coins := sdk.Coins{sdk.NewCoin(initFromCoins[denomIndex].Denom, amt)}
	msg = types.NewMsgSend(fromAcc.Address, toAcc.Address, coins)
	return fromAcc, "", msg, true
}

// Sends and verifies the transition of a msg send.
func sendAndVerifyMsgSend(r *rand.Rand, app *baseapp.BaseApp, ak types.AccountKeeper,
	msg types.MsgSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey) error {
	fromAcc := ak.GetAccount(ctx, msg.FromAddress)
	fees, err := helpers.RandomFees(r, ctx, fromAcc, msg.Amount)
	if err != nil {
		return err
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		fees,
		chainID,
		[]uint64{fromAcc.GetAccountNumber()},
		[]uint64{fromAcc.GetSequence()},
		privkeys...,
	)

	res := app.Deliver(tx)
	if !res.IsOK() {
		return errors.New(res.Log)
	}

	return nil
}

// SimulateSingleInputMsgMultiSend tests and runs a single msg multisend, with one input and one output, where both
// accounts already exist.
func SimulateSingleInputMsgMultiSend(ak types.AccountKeeper, bk keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		fromAcc, comment, msg, ok := createSingleInputMsgMultiSend(r, ctx, accs, ak)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(comment)
		}

		err = sendAndVerifyMsgMultiSend(r, app, ak, msg, ctx, chainID, []crypto.PrivKey{fromAcc.PrivKey})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, ok, comment), nil, nil
	}
}

func createSingleInputMsgMultiSend(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, ak types.AccountKeeper) (
	fromAcc simulation.Account, comment string, msg types.MsgMultiSend, ok bool) {

	fromAcc = simulation.RandomAcc(r, accs)
	toAcc := simulation.RandomAcc(r, accs)

	// Disallow sending money to yourself
	for {
		if !fromAcc.PubKey.Equals(toAcc.PubKey) {
			break
		}
		toAcc = simulation.RandomAcc(r, accs)
	}

	toAddr := toAcc.Address
	initFromCoins := ak.GetAccount(ctx, fromAcc.Address).SpendableCoins(ctx.BlockHeader().Time)

	if len(initFromCoins) == 0 {
		return fromAcc, "skipping, no coins at all", msg, false
	}

	denomIndex := r.Intn(len(initFromCoins))
	amt, err := simulation.RandPositiveInt(r, initFromCoins[denomIndex].Amount)
	if err != nil {
		return fromAcc, "skipping bank send due to account having no coins of denomination " + initFromCoins[denomIndex].Denom, msg, false
	}

	coins := sdk.Coins{sdk.NewCoin(initFromCoins[denomIndex].Denom, amt)}
	msg = types.MsgMultiSend{
		Inputs:  []types.Input{types.NewInput(fromAcc.Address, coins)},
		Outputs: []types.Output{types.NewOutput(toAddr, coins)},
	}

	return fromAcc, "", msg, true
}

// Sends and verifies the transition of a msg multisend. This fails if there are repeated inputs or outputs
// pass in handler as nil to handle txs, otherwise handle msgs
func sendAndVerifyMsgMultiSend(r *rand.Rand, app *baseapp.BaseApp, ak types.AccountKeeper,
	msg types.MsgMultiSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey) error {

	initialInputAddrCoins := make([]sdk.Coins, len(msg.Inputs))
	initialOutputAddrCoins := make([]sdk.Coins, len(msg.Outputs))
	accountNumbers := make([]uint64, len(msg.Inputs))
	sequenceNumbers := make([]uint64, len(msg.Inputs))

	var fees sdk.Coins
	for i := 0; i < len(msg.Inputs); i++ {
		acc := ak.GetAccount(ctx, msg.Inputs[i].Address)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()

		// select a random amount for the transaction
		coins := acc.GetCoins()
		denomIndex := r.Intn(len(coins))
		amt, err := simulation.RandPositiveInt(r, coins[denomIndex].Amount)
		if err != nil {
			continue
		}

		msgAmt := sdk.Coins{sdk.NewCoin(coins[denomIndex].Denom, amt)}
		fee, err := helpers.RandomFees(r, ctx, acc, msgAmt)
		if err != nil {
			continue
		}

		fees = fees.Add(fee)
		initialInputAddrCoins[i] = msgAmt
	}

	for i := 0; i < len(msg.Outputs); i++ {
		acc := ak.GetAccount(ctx, msg.Outputs[i].Address)
		initialOutputAddrCoins[i] = acc.GetCoins()
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		fees,
		chainID,
		accountNumbers,
		sequenceNumbers,
		privkeys...,
	)

	res := app.Deliver(tx)
	if !res.IsOK() {
		return errors.New(res.Log)
	}

	return nil
}
