package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
// will be removed in the future
const (
	OpWeightMsgSend           = "op_weight_msg_send"
	OpWeightMsgMultiSend      = "op_weight_msg_multisend"
	DefaultWeightMsgSend      = 100 // from simappparams.DefaultWeightMsgSend
	DefaultWeightMsgMultiSend = 10  // from simappparams.DefaultWeightMsgMultiSend
)

// WeightedOperations returns all the operations from the module with their respective weights
// migrate to the msg factories instead, this method will be removed in the future
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgSend, weightMsgMultiSend int
	appParams.GetOrGenerate(OpWeightMsgSend, &weightMsgSend, nil, func(_ *rand.Rand) {
		weightMsgSend = DefaultWeightMsgSend
	})

	appParams.GetOrGenerate(OpWeightMsgMultiSend, &weightMsgMultiSend, nil, func(_ *rand.Rand) {
		weightMsgMultiSend = DefaultWeightMsgMultiSend
	})

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSend,
			SimulateMsgSend(txGen, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgMultiSend,
			SimulateMsgMultiSend(txGen, ak, bk),
		),
	}
}

// SimulateMsgSend tests and runs a single msg send where both
// accounts already exist.
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgSend(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgSend{})
		from, to, coins, skip := randomSendFields(r, ctx, accs, bk, ak)

		// if coins slice is empty, we can not create valid types.MsgSend
		if len(coins) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "empty coins slice"), nil, nil
		}

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
		}

		if skip {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "skip all transfers"), nil, nil
		}

		msg := types.NewMsgSend(from.Address, to.Address, coins)

		err := sendMsgSend(r, app, txGen, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgSendToModuleAccount tests and runs a single msg send where both
// accounts already exist.
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgSendToModuleAccount(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk keeper.Keeper,
	moduleAccCount int,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgSend{})
		from := accs[0]
		to := getModuleAccounts(ak, ctx, moduleAccCount)[0]

		spendable := bk.SpendableCoins(ctx, from.Address)
		coins := simtypes.RandSubsetCoins(r, spendable)
		// if coins slice is empty, we can not create valid types.MsgSend
		if len(coins) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, msgType, "empty coins slice"), nil, nil
		}

		// Check send_enabled status of each coin denom
		if err := bk.IsSendEnabledCoins(ctx, coins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
		}

		msg := types.NewMsgSend(from.Address, to.Address, coins)

		err := sendMsgSend(r, app, txGen, bk, ak, msg, ctx, chainID, []cryptotypes.PrivKey{from.PrivKey})
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// sendMsgSend sends a transaction with a MsgSend from a provided random account.
func sendMsgSend(
	r *rand.Rand, app *baseapp.BaseApp,
	txGen client.TxConfig,
	bk keeper.Keeper, ak types.AccountKeeper,
	msg *types.MsgSend, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey,
) error {
	var (
		fees sdk.Coins
		err  error
	)

	from, err := sdk.AccAddressFromBech32(msg.FromAddress)
	if err != nil {
		return err
	}

	account := ak.GetAccount(ctx, from)
	spendable := bk.SpendableCoins(ctx, account.GetAddress())

	coins, hasNeg := spendable.SafeSub(msg.Amount...)
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privkeys...,
	)
	if err != nil {
		return err
	}

	_, _, err = app.SimTxFinalizeBlock(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}

// SimulateMsgMultiSend tests and runs a single msg multisend, with randomized, capped number of inputs/outputs.
// all accounts in msg fields exist in state
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgMultiSend(txGen client.TxConfig, ak types.AccountKeeper, bk keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgMultiSend{})

		// random number of inputs/outputs between [1, 3]
		inputs := make([]types.Input, r.Intn(1)+1) //nolint:staticcheck // SA4030: (*math/rand.Rand).Intn(n) generates a random value 0 <= x < n; that is, the generated values don't include n; r.Intn(1) therefore always returns 0
		outputs := make([]types.Output, r.Intn(3)+1)

		// collect signer privKeys
		privs := make([]cryptotypes.PrivKey, len(inputs))

		// use map to check if address already exists as input
		usedAddrs := make(map[string]bool)

		var totalSentCoins sdk.Coins
		for i := range inputs {
			// generate random input fields, ignore to address
			from, _, coins, skip := randomSendFields(r, ctx, accs, bk, ak)

			// make sure account is fresh and not used in previous input
			for usedAddrs[from.Address.String()] {
				from, _, coins, skip = randomSendFields(r, ctx, accs, bk, ak)
			}

			if skip {
				return simtypes.NoOpMsg(types.ModuleName, msgType, "skip all transfers"), nil, nil
			}

			// set input address in used address map
			usedAddrs[from.Address.String()] = true

			// set signer privkey
			privs[i] = from.PrivKey

			// set next input and accumulate total sent coins
			inputs[i] = types.NewInput(from.Address, coins)
			totalSentCoins = totalSentCoins.Add(coins...)
		}

		// Check send_enabled status of each sent coin denom
		if err := bk.IsSendEnabledCoins(ctx, totalSentCoins...); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
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
				totalSentCoins = totalSentCoins.Sub(outCoins...)
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
		err := sendMsgMultiSend(r, app, txGen, bk, ak, msg, ctx, chainID, privs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgMultiSendToModuleAccount sends coins to Module Accounts
// migrate to the msg factories instead, this method will be removed in the future
func SimulateMsgMultiSendToModuleAccount(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk keeper.Keeper,
	moduleAccCount int,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		msgType := sdk.MsgTypeURL(&types.MsgMultiSend{})
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
			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
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
				totalSentCoins = totalSentCoins.Sub(outCoins...)
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
		err := sendMsgMultiSend(r, app, txGen, bk, ak, msg, ctx, chainID, privs)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// sendMsgMultiSend sends a transaction with a MsgMultiSend from a provided random
// account.
func sendMsgMultiSend(
	r *rand.Rand, app *baseapp.BaseApp,
	txGen client.TxConfig,
	bk keeper.Keeper, ak types.AccountKeeper,
	msg *types.MsgMultiSend, ctx sdk.Context, chainID string, privkeys []cryptotypes.PrivKey,
) error {
	accountNumbers := make([]uint64, len(msg.Inputs))
	sequenceNumbers := make([]uint64, len(msg.Inputs))
	for i := range msg.Inputs {
		addr, err := ak.AddressCodec().StringToBytes(msg.Inputs[i].Address)
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

	addr, err := ak.AddressCodec().StringToBytes(msg.Inputs[0].Address)
	if err != nil {
		panic(err)
	}
	// feePayer is the first signer, i.e. first input address
	feePayer := ak.GetAccount(ctx, addr)
	spendable := bk.SpendableCoins(ctx, feePayer.GetAddress())
	coins, hasNeg := spendable.SafeSub(msg.Inputs[0].Coins...)
	if !hasNeg {
		fees, err = simtypes.RandomFees(r, ctx, coins)
		if err != nil {
			return err
		}
	}
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		fees,
		simtestutil.DefaultGenTxGas,
		chainID,
		accountNumbers,
		sequenceNumbers,
		privkeys...,
	)
	if err != nil {
		return err
	}
	_, _, err = app.SimTxFinalizeBlock(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}
	return nil
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

	for i := range moduleAccCount {
		acc := ak.GetModuleAccount(ctx, disttypes.ModuleName)
		mAcc := simtypes.Account{
			Address: acc.GetAddress(),
			PrivKey: nil,
			ConsKey: nil,
			PubKey:  acc.GetPubKey(),
		}
		moduleAccounts[i] = mAcc
	}

	return moduleAccounts
}
