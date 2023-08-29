package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgSend      = "op_weight_msg_send"
	OpWeightMsgMultiSend = "op_weight_msg_multisend"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk keeper.Keeper,
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
		from, to, coins, skip := randomSendFields(r, ctx, accs, bk, ak)

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, err.Error()), nil, nil
		}

		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgSend(from.Address, to.Address, coins)

		gasInfo, err := sendMsgSend(r, app, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", gasInfo.GasWanted, gasInfo.GasUsed, nil), nil, nil
	}
}

// SimulateMsgSendToModuleAccount tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSendToModuleAccount(ak types.AccountKeeper, bk keeper.Keeper, moduleAccCount int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		from := accs[0]

		to := getModuleAccounts(ak, ctx, moduleAccCount)[0]

		spendable := bk.SpendableCoins(ctx, from.Address)
		coins := simtypes.RandSubsetCoins(r, spendable)

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSend, err.Error()), nil, nil
		}

		msg := types.NewMsgSend(from.Address, to.Address, coins)

		gasInfo, err := sendMsgSend(r, app, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", gasInfo.GasWanted, gasInfo.GasUsed, nil), nil, nil
	}
}

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
func sendMsgSend(
	r *rand.Rand, app *baseapp.BaseApp, bk keeper.Keeper, ak types.AccountKeeper,
	msg *types.MsgSend, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey,
) (sdk.GasInfo, error) {

	var (
		fees sdk.Coins
		err  error
	)

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return sdk.GasInfo{}, err
	}

	account := ak.GetAccount(ctx, from)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Amount)
	if !hasNeg {
		feeCoins := coins.FilterDenoms([]string{sdk.DefaultBondDenom})
		fees, err = simtypes.RandomFees(r, ctx, feeCoins)
		if err != nil {
			return sdk.GasInfo{}, err
		}
	}
	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)
	if err != nil {
		return sdk.GasInfo{}, err
	}

	gasInfo, _, err := app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return sdk.GasInfo{}, err
	}

	return gasInfo, nil
}

// SimulateMsgMultiSend tests and runs a single msg multisend, with randomized, capped number of inputs/outputs.
// all accounts in msg fields exist in state
func SimulateMsgMultiSend(ak types.AccountKeeper, bk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// random number of outputs between [1, 3]
		outputs := make([]types.Output, r.Intn(3)+1)

		// generate random input fields, ignore to address
		from, _, inputCoins, skip := randomSendFields(r, ctx, accs, bk, ak)
		if skip {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMultiSend, "skip all transfers"), nil, nil
		}

		privKeys := []cryptotypes.PrivKey{from.PrivKey}
		inputs := []types.Input{
			types.NewInput(from.Address, inputCoins),
		}

		var totalSentCoins sdk.Coins
		totalSentCoins = totalSentCoins.Add(inputCoins...)

		// check send_enabled status of each sent coin denom
		if err := bk.IsSendEnabledCoins(ctx, inputCoins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMultiSend, err.Error()), nil, nil
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
		for i := 0; i < len(outputs); {
			if outputs[i].Coins.Empty() {
				outputs[i] = outputs[len(outputs)-1]
				outputs = outputs[:len(outputs)-1]
			} else {
				// continue onto next coin
				i++
			}
		}

		msg := &types.MsgMultiSend{
			Inputs:  inputs,
			Outputs: outputs,
		}

		gasInfo, err := sendMsgMultiSend(r, app, bk, ak, msg, ctx, chainID, privKeys)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", gasInfo.GasWanted, gasInfo.GasUsed, nil), nil, nil
	}
}

// SimulateMsgMultiSendToModuleAccount sends coins to Module Accounts
func SimulateMsgMultiSendToModuleAccount(ak types.AccountKeeper, bk keeper.Keeper, moduleAccCount int) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		inputs := make([]types.Input, 2)
		outputs := make([]types.Output, moduleAccCount)
		// collect signer privKeys
		privs := make([]cryptotypes.PrivKey, len(inputs))

		var totalSentCoins sdk.Coins
		for i := range inputs {
			sender := accs[i]
			privs[i] = sender.PrivKey
			spendable := bk.SpendableCoins(ctx, sender.Address)
			coins := simtypes.RandSubsetCoins(r, spendable)
			inputs[i] = types.NewInput(sender.Address, coins)
			totalSentCoins = totalSentCoins.Add(coins...)
		}

		if err := bk.IsSendEnabledCoins(ctx, totalSentCoins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgMultiSend, err.Error()), nil, nil
		}

		moduleAccounts := getModuleAccounts(ak, ctx, moduleAccCount)
		for i := range outputs {
			var outCoins sdk.Coins
			// split total sent coins into random subsets for output
			if i == len(outputs)-1 {
				outCoins = totalSentCoins
			} else {
				// take random subset of remaining coins for output
				// and update remaining coins
				outCoins = simtypes.RandSubsetCoins(r, totalSentCoins)
				totalSentCoins = totalSentCoins.Sub(outCoins)
			}

			outputs[i] = types.NewOutput(moduleAccounts[i].Address, outCoins)
		}

		// remove any output that has no coins

		for i := 0; i < len(outputs); {
			if outputs[i].Coins.Empty() {
				outputs[i] = outputs[len(outputs)-1]
				outputs = outputs[:len(outputs)-1]
			} else {
				// continue onto next coin
				i++
			}
		}

		msg := &types.MsgMultiSend{
			Inputs:  inputs,
			Outputs: outputs,
		}

		gasInfo, err := sendMsgMultiSend(r, app, bk, ak, msg, ctx, chainID, privs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, "", gasInfo.GasWanted, gasInfo.GasUsed, nil), nil, nil
	}
}

// sendMsgMultiSend sends a transaction with a MsgMultiSend from a provided random
// account.
func sendMsgMultiSend(
	r *rand.Rand, app *baseapp.BaseApp, bk keeper.Keeper, ak types.AccountKeeper,
	msg *types.MsgMultiSend, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey,
) (sdk.GasInfo, error) {

	accountNumbers := make([]uint64, len(msg.Inputs))
	sequenceNumbers := make([]uint64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		addr, err := sdk.AccAddressFromBech32(msg.Inputs[i].Address)
		if err != nil {
			panic(err)
		}
		acc := ak.GetAccount(ctx, addr)
		accountNumbers[i] = acc.GetAccountNumber()
		sequenceNumbers[i] = acc.GetSequence()
	}

	var (
		fees sdk.Coins
		err  error
	)

	addr, err := sdk.AccAddressFromBech32(msg.Inputs[0].Address)
	if err != nil {
		return sdk.GasInfo{}, err
	}

	// feePayer is the first signer, i.e. first input address
	feePayer := ak.GetAccount(ctx, addr)
	spendable := bk.SpendableCoins(ctx, feePayer.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Inputs[0].Coins)
	if !hasNeg {
		feeCoins := coins.FilterDenoms([]string{sdk.DefaultBondDenom})
		fees, err = simtypes.RandomFees(r, ctx, feeCoins)
		if err != nil {
			return sdk.GasInfo{}, err
		}
	}

	txGen := simappparams.MakeTestEncodingConfig().TxConfig
	tx, err := helpers.GenTx(
		txGen,
		[]sdk.Msg{msg},
		fees,
		helpers.DefaultGenTxGas,
		chainID,
		accountNumbers,
		sequenceNumbers,
		privkeys...,
	)
	if err != nil {
		return sdk.GasInfo{}, err
	}

	gasInfo, _, err := app.Deliver(txGen.TxEncoder(), tx)
	if err != nil {
		return sdk.GasInfo{}, err
	}

	return gasInfo, nil
}

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
func randomSendFields(
	r *rand.Rand, ctx sdk.Context, accs []simtypes.Account, bk keeper.Keeper, ak types.AccountKeeper,
) (simtypes.Account, simtypes.Account, sdk.Coins, bool) {

	from, _ := simtypes.RandomAcc(r, accs)
	to, _ := simtypes.RandomAcc(r, accs)

	// disallow sending money to yourself
	for from.PubKey.Equals(to.PubKey) {
		to, _ = simtypes.RandomAcc(r, accs)
	}

	acc := ak.GetAccount(ctx, from.Address)
	if acc == nil {
		return from, to, nil, true
	}

	spendable := bk.SpendableCoins(ctx, acc.GetAddress())

	sendCoins := simtypes.RandSubsetCoins(r, spendable)
	if sendCoins.Empty() {
		return from, to, nil, true
	}

	return from, to, sendCoins, false
}

func getModuleAccounts(ak types.AccountKeeper, ctx sdk.Context, moduleAccCount int) []simtypes.Account {
	moduleAccounts := make([]simtypes.Account, moduleAccCount)

	for i := 0; i < moduleAccCount; i++ {
		addr := ak.GetModuleAddress(distributiontypes.ModuleName)
		acc := ak.GetAccount(ctx, addr)

		moduleAccounts[i] = simtypes.Account{
			Address: acc.GetAddress(),
			PrivKey: nil,
			ConsKey: nil,
			PubKey:  acc.GetPubKey(),
		}
	}

	return moduleAccounts
}
