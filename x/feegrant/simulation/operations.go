package simulation

import (
	"context"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation operation weights constants
const (
	OpWeightMsgGrantFeeAllowance  = "op_weight_msg_grant_fee_allowance"
	OpWeightMsgRevokeFeeAllowance = "op_weight_msg_grant_revoke_allowance"
)

var (
	TypeMsgGrantFeeAllowance  = sdk.MsgTypeURL(&types.MsgGrantFeeAllowance{})
	TypeMsgRevokeFeeAllowance = sdk.MsgTypeURL(&types.MsgRevokeFeeAllowance{})
)

func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec,
	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
	protoCdc *codec.ProtoCodec,
) simulation.WeightedOperations {

	var (
		weightMsgGrantFeeAllowance  int
		weightMsgRevokeFeeAllowance int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgGrantFeeAllowance, &weightMsgGrantFeeAllowance, nil,
		func(_ *rand.Rand) {
			weightMsgGrantFeeAllowance = simappparams.DefaultWeightGrantFeeAllowance
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgRevokeFeeAllowance, &weightMsgRevokeFeeAllowance, nil,
		func(_ *rand.Rand) {
			weightMsgRevokeFeeAllowance = simappparams.DefaultWeightRevokeFeeAllowance
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgGrantFeeAllowance,
			SimulateMsgGrantFeeAllowance(ak, bk, k, protoCdc),
		),
		simulation.NewWeightedOperation(
			weightMsgRevokeFeeAllowance,
			SimulateMsgRevokeFeeAllowance(ak, bk, k, protoCdc),
		),
	}
}

// SimulateMsgGrantFeeAllowance generates MsgGrantFeeAllowance with random values.
func SimulateMsgGrantFeeAllowance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		granter, _ := simtypes.RandomAcc(r, accs)
		grantee, _ := simtypes.RandomAcc(r, accs)
		if grantee.Address.String() == granter.Address.String() {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, "grantee and granter cannot be same"), nil, nil
		}

		if f, _ := k.GetFeeAllowance(ctx, granter.Address, grantee.Address); f != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, "fee allowance exists"), nil, nil
		}

		account := ak.GetAccount(ctx, granter.Address)

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, err.Error()), nil, err
		}

		spendableCoins = spendableCoins.Sub(fees)
		if spendableCoins.Empty() {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, "unable to grant empty coins as SpendLimit"), nil, nil
		}

		oneYear := ctx.BlockTime().AddDate(1, 0, 0)
		msg, err := types.NewMsgGrantFeeAllowance(&types.BasicFeeAllowance{
			SpendLimit: spendableCoins,
			Expiration: &oneYear,
		}, granter.Address, grantee.Address)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, err.Error()), nil, err
		}
		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		feegrantMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = feegrantMsgClient.GrantFeeAllowance(context.Background(), msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, err.Error()), nil, err
		}
		tx, err := helpers.GenTx(
			txGen,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			granter.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, "unable to generate mock tx"), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(svcMsgClientConn.GetMsgs()[0]), "unable to deliver tx"), nil, err
		}
		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, err
	}
}

// SimulateMsgRevokeFeeAllowance generates a MsgRevokeFeeAllowance with random values.
func SimulateMsgRevokeFeeAllowance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, protoCdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		hasGrant := false
		var granterAddr sdk.AccAddress
		var granteeAddr sdk.AccAddress
		k.IterateAllFeeAllowances(ctx, func(grant types.FeeAllowanceGrant) bool {

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
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeFeeAllowance, "no grants"), nil, nil
		}
		granter, ok := simtypes.FindAccount(accs, granterAddr)

		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeFeeAllowance, "Account not found"), nil, nil
		}

		account := ak.GetAccount(ctx, granter.Address)
		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeFeeAllowance, err.Error()), nil, err
		}

		msg := types.NewMsgRevokeFeeAllowance(granterAddr, granteeAddr)

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		svcMsgClientConn := &msgservice.ServiceMsgClientConn{}
		feegrantMsgClient := types.NewMsgClient(svcMsgClientConn)
		_, err = feegrantMsgClient.RevokeFeeAllowance(context.Background(), &msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgGrantFeeAllowance, err.Error()), nil, err
		}

		tx, err := helpers.GenTx(
			txGen,
			svcMsgClientConn.GetMsgs(),
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			granter.PrivKey,
		)

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, TypeMsgRevokeFeeAllowance, err.Error()), nil, err
		}

		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
		return simtypes.NewOperationMsg(svcMsgClientConn.GetMsgs()[0], true, "", protoCdc), nil, err
	}
}
