package simulation

import (
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgSend      = "op_weight_msg_send"
	OpWeightMsgMultiSend = "op_weight_msg_multisend"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc *codec.Codec, ak types.AccountKeeper, bk keeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgSend, weightMsgMultiSend int
	appParams.GetOrGenerate(cdc, OpWeightMsgSend, &weightMsgSend, nil,
		func(_ *rand.Rand) {
			weightMsgSend = simappparams.DefaultWeightMsgSend
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgMultiSend, &weightMsgMultiSend, nil,
		func(_ *rand.Rand) {
			weightMsgMultiSend = simappparams.DefaultWeightMsgMultiSend
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSend,
			SimulateMsgSend(ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgMultiSend,
			SimulateMsgMultiSend(ak, bk),
		),
	}
}

// SimulateMsgSend tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSend(ak types.AccountKeeper, bk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		if !bk.GetSendEnabled(ctx) {
			return simtypes.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, toSimAcc, coins, skip, err := randomSendFields(r, ctx, accs, bk, ak)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName), nil, err
		}

		if skip {
			return simtypes.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgSend(simAccount.Address, toSimAcc.Address, coins)

		err = sendMsgSend(r, app, bk, ak, msg, ctx, chainID, []crypto.PrivKey{simAccount.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
// nolint: interfacer
func sendMsgSend(
	r *rand.Rand, app *baseapp.BaseApp, bk keeper.Keeper, ak types.AccountKeeper,
	msg types.MsgSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey,
) error {

	var (
		fees sdk.Coins
		err  error
	)

	account := ak.GetAccount(ctx, msg.FromAddress)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Amount)
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)

	_, _, err = app.Deliver(tx)
	if err != nil {
		return err
	}

	return nil
}

// SimulateMsgMultiSend tests and runs a single msg multisend, with randomized, capped number of inputs/outputs.
// all accounts in msg fields exist in state
func SimulateMsgMultiSend(ak types.AccountKeeper, bk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		if !bk.GetSendEnabled(ctx) {
			return simtypes.NoOpMsg(types.ModuleName), nil, nil
		}

		// random number of inputs/outputs between [1, 3]
		inputs := make([]types.Input, r.Intn(3)+1)
		outputs := make([]types.Output, r.Intn(3)+1)

		// collect signer privKeys
		privs := make([]crypto.PrivKey, len(inputs))

		// use map to check if address already exists as input
		usedAddrs := make(map[string]bool)

		var totalSentCoins sdk.Coins
		for i := range inputs {
			// generate random input fields, ignore to address
			simAccount, _, coins, skip, err := randomSendFields(r, ctx, accs, bk, ak)

			// make sure account is fresh and not used in previous input
			for usedAddrs[simAccount.Address.String()] {
				simAccount, _, coins, skip, err = randomSendFields(r, ctx, accs, bk, ak)
			}

			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName), nil, err
			}
			if skip {
				return simtypes.NoOpMsg(types.ModuleName), nil, nil
			}

			// set input address in used address map
			usedAddrs[simAccount.Address.String()] = true

			// set signer privkey
			privs[i] = simAccount.PrivKey

			// set next input and accumulate total sent coins
			inputs[i] = types.NewInput(simAccount.Address, coins)
			totalSentCoins = totalSentCoins.Add(coins...)
		}

		for o := range outputs {
			outAddr, _ := simtypes.RandomAcc(r, accs)

			var outCoins sdk.Coins
			// split total sent coins into random subsets for output
			if o == len(outputs)-1 {
				outCoins = totalSentCoins
			} else {
				// take random subset of remaining coins for output
				// and update remaining coins
				outCoins = simtypes.RandSubsetCoins(r, totalSentCoins)
				totalSentCoins = totalSentCoins.Sub(outCoins)
			}

			outputs[o] = types.NewOutput(outAddr.Address, outCoins)
		}

		// remove any output that has no coins
		i := 0
		for i < len(outputs) {
			if outputs[i].Coins.Empty() {
				outputs[i] = outputs[len(outputs)-1]
				outputs = outputs[:len(outputs)-1]
			} else {
				// continue onto next coin
				i++
			}
		}

		msg := types.MsgMultiSend{
			Inputs:  inputs,
			Outputs: outputs,
		}

		err := sendMsgMultiSend(r, app, bk, ak, msg, ctx, chainID, privs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// sendMsgMultiSend sends a transaction with a MsgMultiSend from a provided random
// account.
// nolint: interfacer
func sendMsgMultiSend(
	r *rand.Rand, app *baseapp.BaseApp, bk keeper.Keeper, ak types.AccountKeeper,
	msg types.MsgMultiSend, ctx sdk.Context, chainID string, privkeys []crypto.PrivKey,
) error {

	accountNumbers := make([]uint64, len(msg.Inputs))
	sequenceNumbers := make([]uint64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		acc := ak.GetAccount(ctx, msg.Inputs[i].Address)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()
	}

	var (
		fees sdk.Coins
		err  error
	)

	// feePayer is the first signer, i.e. first input address
	feePayer := ak.GetAccount(ctx, msg.Inputs[0].Address)
	spendable := bk.SpendableCoins(ctx, feePayer.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Inputs[0].Coins)
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}

	tx := helpers.GenTx(
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		accountNumbers,
		sequenceNumbers,
		privkeys...,
	)

	_, _, err = app.Deliver(tx)
	if err != nil {
		return err
	}

	return nil
}

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
// nolint: interfacer
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk keeper.Keeper, ak types.AccountKeeper,
) (simtypes.Account, simtypes.Account, sdk.Coins, bool, error) {

	simAccount, _ := simtypes.RandomAcc(r, accs)
	toSimAcc, _ := simtypes.RandomAcc(r, accs)

	// disallow sending money to yourself
	for simAccount.PubKey.Equals(toSimAcc.PubKey) {
		toSimAcc, _ = simtypes.RandomAcc(r, accs)
	}

	acc := ak.GetAccount(ctx, simAccount.Address)
	if acc == nil {
		return simAccount, toSimAcc, nil, true, nil // skip error
	}

	spendable := bk.SpendableCoins(ctx, acc.GetAddress())

	sendCoins := simtypes.RandSubsetCoins(r, spendable)
	if sendCoins.Empty() {
		return simAccount, toSimAcc, nil, true, nil // skip error
	}

	return simAccount, toSimAcc, sendCoins, false, nil
}
