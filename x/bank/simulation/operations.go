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
func SimulateMsgSend(ak types.AccountKeeper, bk keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		if !bk.GetSendEnabled(ctx) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, msg, skip, err := createMsgSend(r, ctx, accs, ak)
		switch {
		case skip:
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case err != nil:
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		err = sendMsgSend(r, app, ak, msg, ctx, chainID, []crypto.PrivKey{simAccount.PrivKey})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func createMsgSend(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, ak types.AccountKeeper) (
	simAccount simulation.Account, msg types.MsgSend, skip bool, err error) {

	simAccount, _ = simulation.RandomAcc(r, accs)
	toSimAcc, idx := simulation.RandomAcc(r, accs)

	// Disallow sending money to yourself
	var accsCopy []simulation.Account
	accsCopy = append(accsCopy, accs...)
	for len(accsCopy) > 0 {
		if !simAccount.PubKey.Equals(toSimAcc.PubKey) {
			break
		}

		accsCopy = append(accsCopy[:idx], accsCopy[idx+1:]...)
		toSimAcc, idx = simulation.RandomAcc(r, accsCopy)
	}

	if len(accsCopy) == 0 {
		return simAccount, msg, false, errors.New("all accounts are equal")
	}

	acc := ak.GetAccount(ctx, simAccount.Address)
	if acc == nil {
		return simAccount, msg, true, nil
	}

	coins := acc.SpendableCoins(ctx.BlockHeader().Time)
	if coins.Empty() {
		return simAccount, msg, true, nil
	}

	denomIndex := r.Intn(len(coins))
	if coins[denomIndex].Amount.IsZero() {
		return simAccount, msg, true, nil
	}

	amt, err := simulation.RandPositiveInt(r, coins[denomIndex].Amount)
	if err != nil {
		return simAccount, msg, false, err
	}

	coins = sdk.Coins{sdk.NewCoin(coins[denomIndex].Denom, amt)}
	msg = types.NewMsgSend(simAccount.Address, toSimAcc.Address, coins)
	return simAccount, msg, false, nil
}

// Sends and verifies the transition of a msg send.
func sendMsgSend(r *rand.Rand, app *baseapp.BaseApp, ak types.AccountKeeper,
	msg types.MsgSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey) error {
	simAccount := ak.GetAccount(ctx, msg.FromAddress)
	fees, err := helpers.RandomFees(r, ctx, simAccount, msg.Amount)
	if err != nil {
		return err
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		fees,
		chainID,
		[]uint64{simAccount.GetAccountNumber()},
		[]uint64{simAccount.GetSequence()},
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

		if !bk.GetSendEnabled(ctx) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, msg, skip, err := createSingleInputMsgMultiSend(r, ctx, accs, ak)
		switch {
		case skip:
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case err != nil:
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		err = sendMsgMultiSend(r, app, ak, msg, ctx, chainID, []crypto.PrivKey{simAccount.PrivKey})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

func createSingleInputMsgMultiSend(r *rand.Rand, ctx sdk.Context, accs []simulation.Account, ak types.AccountKeeper) (
	simAccount simulation.Account, msg types.MsgMultiSend, skip bool, err error) {

	simAccount, _ = simulation.RandomAcc(r, accs)
	toSimAcc, idx := simulation.RandomAcc(r, accs)

	// Disallow sending money to yourself
	var accsCopy []simulation.Account
	accsCopy = append(accsCopy, accs...)
	for len(accsCopy) > 0 {
		if !simAccount.PubKey.Equals(toSimAcc.PubKey) {
			break
		}

		accsCopy = append(accsCopy[:idx], accsCopy[idx+1:]...)
		toSimAcc, idx = simulation.RandomAcc(r, accsCopy)
	}

	if len(accsCopy) == 0 {
		return simAccount, msg, false, errors.New("all accounts are equal")
	}

	acc := ak.GetAccount(ctx, simAccount.Address)
	if acc == nil {
		return simAccount, msg, true, nil
	}

	coins := acc.SpendableCoins(ctx.BlockHeader().Time)
	if coins.Empty() {
		return simAccount, msg, true, nil // skip without returning any error
	}

	denomIndex := r.Intn(len(coins))
	if coins[denomIndex].Amount.IsZero() {
		return simAccount, msg, true, nil
	}

	amt, err := simulation.RandPositiveInt(r, coins[denomIndex].Amount)
	if err != nil {
		return simAccount, msg, false, nil
	}

	coins = sdk.Coins{sdk.NewCoin(coins[denomIndex].Denom, amt)}
	msg = types.MsgMultiSend{
		Inputs:  []types.Input{types.NewInput(simAccount.Address, coins)},
		Outputs: []types.Output{types.NewOutput(toSimAcc.Address, coins)},
	}

	return simAccount, msg, false, nil
}

// Sends and verifies the transition of a msg multisend. This fails if there are repeated inputs or outputs
// pass in handler as nil to handle txs, otherwise handle msgs
func sendMsgMultiSend(r *rand.Rand, app *baseapp.BaseApp, ak types.AccountKeeper,
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
		coins := acc.SpendableCoins(ctx.BlockHeader().Time)
		denomIndex := r.Intn(len(coins))
		if coins[denomIndex].Amount.IsZero() {
			// skip
			continue
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
		if i == 0 {
			var err error
			fees, err = helpers.RandomFees(r, ctx, acc, initialOutputAddrCoins[i])
			if err != nil {
				return err
			}
		}
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
