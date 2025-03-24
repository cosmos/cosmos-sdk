package simulation

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/keeper"
	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"math/rand"
)

// Governance message types and routes
var (
	TypeUpdateParams          = sdk.MsgTypeURL(&types.MsgUpdateParams{})
	TypeFundCommunityPool     = sdk.MsgTypeURL(&types.MsgFundCommunityPool{})
	TypeMsgCommunityPoolSpend = sdk.MsgTypeURL(&types.MsgCommunityPoolSpend{})
	TypeCreateContinuousFund  = sdk.MsgTypeURL(&types.MsgCreateContinuousFund{})
	TypeCancelContinuousFund  = sdk.MsgTypeURL(&types.MsgCancelContinuousFund{})
)

// Simulation operation weights constants
const (
	OpWeightMsgUpdateParams         = "op_weight_msg_update_params"
	OpWeightMsgCommunityPoolSpend   = "op_weight_msg_community_pool_spend"
	OpWeightMsgFundCommunityPool    = "op_weight_msg_fund_community_pool"
	OpWeightMsgCancelContinuousFund = "op_weight_msg_cancel_continuous_fund"
	OpWeightMsgCreateContinuousFund = "op_weight_msg_create_continuous_fund"

	DefaultWeightMsgFundCommunityPool    = 100
	DefaultWeightMsgCreateContinuousFund = 33
	DefaultWeightMsgCancelContinuousFund = 33
	DefaultWeightMsgCommunityPoolSpend   = 5
	DefaultWeightMsgUpdateParams         = 5
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgUpdateParams         int
		weightMsgCommunityPoolSpend   int
		weightMsgFundCommunityPool    int
		weightMsgCreateContinuousFund int
		weightMsgCancelContinuousFund int
	)

	appParams.GetOrGenerate(OpWeightMsgUpdateParams, &weightMsgUpdateParams, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateParams = DefaultWeightMsgUpdateParams
		},
	)

	appParams.GetOrGenerate(OpWeightMsgCommunityPoolSpend, &weightMsgCommunityPoolSpend, nil,
		func(_ *rand.Rand) {
			weightMsgCommunityPoolSpend = DefaultWeightMsgCommunityPoolSpend
		},
	)

	appParams.GetOrGenerate(OpWeightMsgFundCommunityPool, &weightMsgFundCommunityPool, nil,
		func(_ *rand.Rand) {
			weightMsgFundCommunityPool = DefaultWeightMsgFundCommunityPool
		},
	)

	appParams.GetOrGenerate(OpWeightMsgCreateContinuousFund, &weightMsgCreateContinuousFund, nil,
		func(_ *rand.Rand) {
			weightMsgCreateContinuousFund = DefaultWeightMsgCreateContinuousFund
		},
	)

	appParams.GetOrGenerate(OpWeightMsgCancelContinuousFund, &weightMsgCancelContinuousFund, nil,
		func(_ *rand.Rand) {
			weightMsgCancelContinuousFund = DefaultWeightMsgCancelContinuousFund
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgUpdateParams,
			SimulateMsgUpdateParams(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgFundCommunityPool,
			SimulateMsgFundCommunityPool(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCommunityPoolSpend,
			SimulateMsgCommunityPoolSpend(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateContinuousFund,
			SimulateMsgCreateContinuousFund(txGen, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCancelContinuousFund,
			SimulateMsgCancelContinuousFund(txGen, ak, bk, k),
		),
	}
}

func SimulateMsgUpdateParams(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// generate an random k 1-100

		return simtypes.OperationMsg{}, nil, nil
	}
}

func SimulateMsgCommunityPoolSpend(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// find random account and send random balance in pool from auth

		return simtypes.OperationMsg{}, nil, nil
	}
}

func SimulateMsgFundCommunityPool(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// find random acct with balance

		return simtypes.OperationMsg{}, nil, nil
	}
}

func SimulateMsgCreateContinuousFund(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// generate random percent and assign to random account

		return simtypes.OperationMsg{}, nil, nil
	}
}

func SimulateMsgCancelContinuousFund(
	txGen client.TxConfig,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		// find a random continuous fund in state and cancel

		return simtypes.OperationMsg{}, nil, nil
	}
}
