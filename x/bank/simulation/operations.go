package simulation

import (
	"math/rand"

	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgSend           = "op_weight_msg_send"
	OpWeightMsgMultiSend      = "op_weight_msg_multisend"
	DefaultWeightMsgSend      = 100 // from simappparams.DefaultWeightMsgSend
	DefaultWeightMsgMultiSend = 10  // from simappparams.DefaultWeightMsgMultiSend

	distributionModuleName = "distribution"
)

// WeightedOperations returns all the operations from the module with their respective weights
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
	reporter := &BasicSimulationReporter{}
	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSend,
			SimulateMsgSend(reporter, txGen, ak, bk),
		),
		simulation.NewWeightedOperation(
			weightMsgMultiSend,
			SimulateMsgMultiSend(reporter, txGen, ak, bk),
		),
	}
}

// SimulateMsgSend tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSend(
	reporter SimulationReporter,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		testData := NewChainDataSource(r, ak, NewBalanceSource(ctx, bk), ak.AddressCodec(), accs...)
		reporter = reporter.WithScope(&types.MsgSend{})

		sender, msg := MsgSendFactory(bk, testData, reporter, ctx)
		return DeliverSimsMsg(reporter, r, app, txGen, ak, msg, ctx, chainID, sender...), nil, reporter.ExecutionResult().Error
	}
}

// SimulateMsgSendToModuleAccount tests and runs a single msg send where both
// accounts already exist.
func SimulateMsgSendToModuleAccount(
	testData *ChainDataSource,
	reporter SimulationReporter,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		testData := NewChainDataSource(r, ak, NewBalanceSource(ctx, bk), ak.AddressCodec(), accs...)
		reporter = reporter.WithScope(&types.MsgSend{})

		from, msg := MsgSendToModuleAccountFactory(bk, testData, reporter, ctx)
		return DeliverSimsMsg(reporter, r, app, txGen, ak, msg, ctx, chainID, from...), nil, reporter.ExecutionResult().Error
	}
}

// SimulateMsgMultiSend tests and runs a single msg multisend, with randomized, capped number of inputs/outputs.
// all accounts in msg fields exist in state
func SimulateMsgMultiSend(
	reporter SimulationReporter,
	txGen client.TxConfig, ak types.AccountKeeper, bk keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		reporter = reporter.WithScope(&types.MsgMultiSend{})
		testData := NewChainDataSource(r, ak, NewBalanceSource(ctx, bk), ak.AddressCodec(), accs...)

		senders, msg := MsgMultiSendFactory(bk, testData, reporter, ctx)
		return DeliverSimsMsg(reporter, r, app, txGen, ak, msg, ctx, chainID, senders...), nil, reporter.ExecutionResult().Error
	}
}

// Alex: not porting this for now as it is very similar to the MsgMultiSend and was not active anyway
//
// SimulateMsgMultiSendToModuleAccount sends coins to Module Accounts
//func SimulateMsgMultiSendToModuleAccount(
//	txGen client.TxConfig,
//	ak types.AccountKeeper,
//	bk keeper.Keeper,
//	moduleAccount int,
//) simtypes.Operation {
//	return func(
//		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
//		accs []simtypes.Account, chainID string,
//	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
//		msgType := sdk.MsgTypeURL(&types.MsgMultiSend{})
//		inputs := make([]types.Input, 2)
//		outputs := make([]types.Output, moduleAccount)
//		// collect signer privKeys
//		privs := make([]cryptotypes.PrivKey, len(inputs))
//		var totalSentCoins sdk.Coins
//		for i := range inputs {
//			sender := accs[i]
//			privs[i] = sender.PrivKey
//			senderAddr, err := ak.AddressCodec().BytesToString(sender.Address)
//			if err != nil {
//				return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, err
//			}
//			spendable := bk.SpendableCoins(ctx, sender.Address)
//			coins := simtypes.RandSubsetCoins(r, spendable)
//			inputs[i] = types.NewInput(senderAddr, coins)
//			totalSentCoins = totalSentCoins.Add(coins...)
//		}
//		if err := bk.IsSendEnabledCoins(ctx, totalSentCoins...); err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, msgType, err.Error()), nil, nil
//		}
//		moduleAccounts := getModuleAccounts(ak, ctx, moduleAccount)
//		for i := range outputs {
//			outAddr, err := ak.AddressCodec().BytesToString(moduleAccounts[i].Address)
//			if err != nil {
//				return simtypes.NoOpMsg(types.ModuleName, msgType, "could not retrieve output address"), nil, err
//			}
//
//			var outCoins sdk.Coins
//			// split total sent coins into random subsets for output
//			if i == len(outputs)-1 {
//				outCoins = totalSentCoins
//			} else {
//				// take random subset of remaining coins for output
//				// and update remaining coins
//				outCoins = simtypes.RandSubsetCoins(r, totalSentCoins)
//				totalSentCoins = totalSentCoins.Sub(outCoins...)
//			}
//			outputs[i] = types.NewOutput(outAddr, outCoins)
//		}
//		// remove any output that has no coins
//		for i := 0; i < len(outputs); {
//			if outputs[i].Coins.Empty() {
//				outputs[i] = outputs[len(outputs)-1]
//				outputs = outputs[:len(outputs)-1]
//			} else {
//				// continue onto next coin
//				i++
//			}
//		}
//		msg := &types.MsgMultiSend{
//			Inputs:  inputs,
//			Outputs: outputs,
//		}
//		err := sendMsgMultiSend(r, app, txGen, bk, ak, msg, ctx, chainID, privs)
//		if err != nil {
//			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid transfers"), nil, err
//		}
//		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
//	}
//}

// randomSendFields returns the sender and recipient simulation accounts as well
// as the transferred amount.
