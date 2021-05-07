package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgGrantAllowance  = "op_weight_msg_grant_fee_allowance"
	OpWeightMsgRevokeAllowance = "op_weight_msg_grant_revoke_allowance"
)

var (
	TypeMsgGrantAllowance  = sdk.MsgTypeURL(&feegrant.MsgGrantAllowance{})
	TypeMsgRevokeAllowance = sdk.MsgTypeURL(&feegrant.MsgRevokeAllowance{})
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak feegrant.AccountKeeper, bk feegrant.BankKeeper, k keeper.Keeper,
	protoCdc *codec.ProtoCodec,
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
			SimulateMsgGrantAllowance(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightMsgRevokeAllowance,
			SimulateMsgRevokeAllowance(ak, bk, k, protoCdc),
		),
	}
}

// SimulateMsgGrantAllowance generates MsgGrantAllowance with random values.
func SimulateMsgGrantAllowance(ak feegrant.AccountKeeper, bk feegrant.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
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
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgGrantAllowance, err.Error()), nil, err
		}

		spendableCoins = spendableCoins.Sub(fees)
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
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenerateTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			granter.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgGrantAllowance, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)

		if err != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, sdk.MsgTypeURL(msg), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(msg, true, "", protoCdc), nil, err
	}
}

// SimulateMsgRevokeAllowance generates a MsgRevokeAllowance with random values.
func SimulateMsgRevokeAllowance(ak feegrant.AccountKeeper, bk feegrant.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		hasGrant := false
		var granterAddr sdk.AccAddress
		var granteeAddr sdk.AccAddress
		k.IterateAllFeeAllowances(ctx, func(grant feegrant.Grant) bool {

			granter, err := sdk.AccAddressFromBech32(grant.Granter)
			if err != nil {
				panic(err)
			}
			grantee, err := sdk.AccAddressFromBech32(grant.Grantee)
			if err != nil {
				panic(err)
			}
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
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgRevokeAllowance, err.Error()), nil, err
		}

		msg := feegrant.NewMsgRevokeAllowance(granterAddr, granteeAddr)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenerateTx(
			txGen,
			[]sdk.Msg{&msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			granter.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(feegrant.ModuleName, TypeMsgRevokeAllowance, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		return simtypes.NewOperationMsg(&msg, true, "", protoCdc), nil, err
	}
}
