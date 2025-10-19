// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"cosmossdk.io/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/timechain/keeper"
	"github.com/cosmos/cosmos-sdk/x/timechain/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgProposeSlot = "op_weight_msg_propose_slot"
	OpWeightMsgConfirmSlot = "op_weight_msg_confirm_slot"
	OpWeightMsgRelayEvent  = "op_weight_msg_relay_event"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgProposeSlot int
		weightMsgConfirmSlot int
		weightMsgRelayEvent  int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgProposeSlot, &weightMsgProposeSlot, nil,
		func(_ *rand.Rand) {
			weightMsgProposeSlot = 100
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgConfirmSlot, &weightMsgConfirmSlot, nil,
		func(_ *rand.Rand) {
			weightMsgConfirmSlot = 100
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgRelayEvent, &weightMsgRelayEvent, nil,
		func(_ *rand.Rand) {
			weightMsgRelayEvent = 100
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgProposeSlot,
			SimulateMsgProposeSlot(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgConfirmSlot,
			SimulateMsgConfirmSlot(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRelayEvent,
			SimulateMsgRelayEvent(ak, bk, k),
		),
	}
}

// SimulateMsgProposeSlot generates a MsgProposeSlot with random values
func SimulateMsgProposeSlot(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// TODO: implement
		return simtypes.OperationMsg{}, nil, nil
	}
}

// SimulateMsgConfirmSlot generates a MsgConfirmSlot with random values
func SimulateMsgConfirmSlot(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// TODO: implement
		return simtypes.OperationMsg{}, nil, nil
	}
}

// SimulateMsgRelayEvent generates a MsgRelayEvent with random values
func SimulateMsgRelayEvent(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// TODO: implement
		return simtypes.OperationMsg{}, nil, nil
	}
}
