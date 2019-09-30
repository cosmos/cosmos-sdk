package simulation

import (
	"errors"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgSend                 = "op_weight_msg_send"
	OpWeightSingleInputMsgMultiSend = "op_weight_single_input_msg_multisend"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	bk keeper.Keeper) simulation.WeightedOperations {

	var weightMsgSend, weightSingleInputMsgMultiSend int
	appParams.GetOrGenerate(cdc, OpWeightMsgSend, &weightMsgSend, nil,
		func(_ *rand.Rand) { weightMsgSend = 100 })

	appParams.GetOrGenerate(cdc, OpWeightSingleInputMsgMultiSend, &weightSingleInputMsgMultiSend, nil,
		func(_ *rand.Rand) { weightSingleInputMsgMultiSend = 10 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgSend,
			SimulateMsgSend(ak, bk),
		),
		simulation.NewWeigthedOperation(
			weightSingleInputMsgMultiSend,
			SimulateSingleInputMsgMultiSend(ak, bk),
		),
	}
}

// SimulateMsgSend tests and runs a single msg send where both
// accounts already exist.
// nolint: funlen
func SimulateMsgSend(ak types.AccountKeeper, bk keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {
		if !bk.GetSendEnabled(ctx) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, toSimAcc, coins, skip, err := randomSendFields(r, ctx, accs, ak)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if skip {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgSend(simAccount.Address, toSimAcc.Address, coins)

		err = sendMsgSend(r, app, ak, msg, ctx, chainID, []crypto.PrivKey{simAccount.PrivKey})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
func sendMsgSend(r *rand.Rand, app *baseapp.BaseApp, ak types.AccountKeeper,
	msg types.MsgSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey) (err error) {
	account := ak.GetAccount(ctx, msg.FromAddress)
	coins := account.SpendableCoins(ctx.BlockTime())

	var fees sdk.Coins
	coins, hasNeg := coins.SafeSub(msg.Amount)
	if !hasNeg {
		fees, err = simulation.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		fees,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
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
// nolint: funlen
func SimulateSingleInputMsgMultiSend(ak types.AccountKeeper, bk keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		simulation.OperationMsg, []simulation.FutureOperation, error) {
		if !bk.GetSendEnabled(ctx) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, toSimAcc, coins, skip, err := randomSendFields(r, ctx, accs, ak)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if skip {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.MsgMultiSend{
			Inputs:  []types.Input{types.NewInput(simAccount.Address, coins)},
			Outputs: []types.Output{types.NewOutput(toSimAcc.Address, coins)},
		}

		err = sendMsgMultiSend(r, app, ak, msg, ctx, chainID, []crypto.PrivKey{simAccount.PrivKey})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// sendMsgMultiSend sends a transaction with a MsgMultiSend from a provided random
// account.
func sendMsgMultiSend(r *rand.Rand, app *baseapp.BaseApp, ak types.AccountKeeper,
	msg types.MsgMultiSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey) (err error) {
	initialInputAddrCoins := make([]sdk.Coins, len(msg.Inputs))
	initialOutputAddrCoins := make([]sdk.Coins, len(msg.Outputs))
	accountNumbers := make([]uint64, len(msg.Inputs))
	sequenceNumbers := make([]uint64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		acc := ak.GetAccount(ctx, msg.Inputs[i].Address)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()

		// select a random amount for the transaction
		coins := acc.SpendableCoins(ctx.BlockHeader().Time)
		denomIndex := r.Intn(len(coins))
		if coins[denomIndex].Amount.IsZero() {
			continue // skip
		}

		amt, err := simulation.RandPositiveInt(r, coins[denomIndex].Amount)
		if err != nil {
			return err
		}

		initialInputAddrCoins[i] = sdk.Coins{sdk.NewCoin(coins[denomIndex].Denom, amt)}
	}

	for i := 0; i < len(msg.Outputs); i++ {
		acc := ak.GetAccount(ctx, msg.Outputs[i].Address)
		initialOutputAddrCoins[i] = acc.SpendableCoins(ctx.BlockHeader().Time)
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		nil, // zero fees
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

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
func randomSendFields(r *rand.Rand, ctx sdk.Context, accs []simulation.Account,
	ak types.AccountKeeper) (simulation.Account, simulation.Account, sdk.Coins, bool, error) {
	simAccount, _ := simulation.RandomAcc(r, accs)
	toSimAcc, _ := simulation.RandomAcc(r, accs)

	// disallow sending money to yourself
	for simAccount.PubKey.Equals(toSimAcc.PubKey) {
		toSimAcc, _ = simulation.RandomAcc(r, accs)
	}

	acc := ak.GetAccount(ctx, simAccount.Address)
	if acc == nil {
		return simAccount, toSimAcc, nil, true, nil // skip error
	}

	coins := acc.SpendableCoins(ctx.BlockHeader().Time)
	if coins.Empty() {
		return simAccount, toSimAcc, nil, true, nil // skip error
	}

	denomIndex := r.Intn(len(coins))
	if coins[denomIndex].Amount.IsZero() {
		return simAccount, toSimAcc, nil, true, nil // skip error
	}

	amt, err := simulation.RandPositiveInt(r, coins[denomIndex].Amount)
	if err != nil {
		return simAccount, toSimAcc, nil, false, err
	}

	coins = sdk.Coins{sdk.NewCoin(coins[denomIndex].Denom, amt)}
	return simAccount, toSimAcc, coins, false, nil
}
