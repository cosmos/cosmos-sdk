package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgGrantAllowance  = "op_weight_msg_grant_fee_allowance"    //nolint:gosec
	OpWeightMsgRevokeAllowance = "op_weight_msg_grant_revoke_allowance" //nolint:gosec
)

var (
	TypeMsgGrantAllowance  = sdk.MsgTypeURL(&feegrant.MsgGrantAllowance{})
	TypeMsgRevokeAllowance = sdk.MsgTypeURL(&feegrant.MsgRevokeAllowance{})
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak feegrant.AccountKeeper, bk feegrant.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgGrantAllowance  int
		weightMsgRevokeAllowance int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgGrantAllowance, &weightMsgGrantAllowance, nil,
		func(_ *rand.Rand) {
			weightMsgGrantAllowance = simappparams.DefaultWeightGrantAllowance
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgRevokeAllowance, &weightMsgRevokeAllowance, nil,
		func(_ *rand.Rand) {
			weightMsgRevokeAllowance = simappparams.DefaultWeightRevokeAllowance
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgGrantAllowance,
			SimulateMsgGrantAllowance(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRevokeAllowance,
			SimulateMsgRevokeAllowance(ak, bk, k),
		),
	}
}

// SimulateMsgGrantAllowance generates MsgGrantAllowance with random values.
func SimulateMsgGrantAllowance(ak feegrant.AccountKeeper, bk feegrant.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		granter, _ := simtypes.RandomAcc(r, accs)
		grantee, _ := simtypes.RandomAcc(r, accs)
		if grantee.Address.String() == granter.Address.String() {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgGrantAllowance, "grantee and granter cannot be same"), nil, nil
		}

		if f, _ := k.GetAllowance(ctx, granter.Address, grantee.Address); f != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgGrantAllowance, "fee allowance exists"), nil, nil
		}

		account := ak.GetAccount(ctx, granter.Address)

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		if spendableCoins.Empty() {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgGrantAllowance, "unable to grant empty coins as SpendLimit"), nil, nil
		}

		oneYear := ctx.BlockTime().AddDate(1, 0, 0)
		msg, err := feegrant.NewMsgGrantAllowance(&feegrant.BasicAllowance{
			SpendLimit: spendableCoins,
			Expiration: &oneYear,
		}, granter.Address, grantee.Address)
		if err != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgGrantAllowance, err.Error()), nil, err
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         TypeMsgGrantAllowance,
			Context:         ctx,
			SimAccount:      granter,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      feegrant.ModuleName,
			CoinsSpentInMsg: spendableCoins,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgRevokeAllowance generates a MsgRevokeAllowance with random values.
func SimulateMsgRevokeAllowance(ak feegrant.AccountKeeper, bk feegrant.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		hasGrant := false
		var granterAddr sdk.AccAddress
		var granteeAddr sdk.AccAddress
		k.IterateAllFeeAllowances(ctx, func(grant feegrant.Grant) bool {
			granter := sdk.MustAccAddressFromBech32(grant.Granter)
			grantee := sdk.MustAccAddressFromBech32(grant.Grantee)
			granterAddr = granter
			granteeAddr = grantee
			hasGrant = true
			return true
		})

		if !hasGrant {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgRevokeAllowance, "no grants"), nil, nil
		}
		granter, ok := simtypes.FindAccount(accs, granterAddr)

		if !ok {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgRevokeAllowance, "Account not found"), nil, nil
		}

		account := ak.GetAccount(ctx, granter.Address)
		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())

		msg := feegrant.NewMsgRevokeAllowance(granterAddr, granteeAddr)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             &msg,
			MsgType:         TypeMsgRevokeAllowance,
			Context:         ctx,
			SimAccount:      granter,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      feegrant.ModuleName,
			CoinsSpentInMsg: spendableCoins,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
