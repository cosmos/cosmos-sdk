package simulation

// import (
// 	"math/rand"
// 	"time"

// 	"github.com/cosmos/cosmos-sdk/baseapp"
// 	"github.com/cosmos/cosmos-sdk/codec"
// 	"github.com/cosmos/cosmos-sdk/simapp/helpers"
// 	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
// 	"github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
// 	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
// 	"github.com/cosmos/cosmos-sdk/x/simulation"
// )

// // Simulation operation weights constants
// const (
// 	OpWeightMsgGrantFeeAllowance  = "op_weight_msg_grant_fee_allowance"
// 	OpWeightMsgRevokeFeeAllowance = "op_weight_msg_grant_revoke_allowance"
// )

// func WeightedOperations(
// 	appParams simtypes.AppParams, cdc codec.JSONMarshaler,
// 	ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper,
// ) simulation.WeightedOperations {

// 	var (
// 		weightMsgGrantFeeAllowance  int
// 		weightMsgRevokeFeeAllowance int
// 	)

// 	appParams.GetOrGenerate(cdc, OpWeightMsgGrantFeeAllowance, &weightMsgGrantFeeAllowance, nil,
// 		func(_ *rand.Rand) {
// 			weightMsgGrantFeeAllowance = simappparams.DefaultWeightGrantFeeAllowance
// 		},
// 	)

// 	appParams.GetOrGenerate(cdc, OpWeightMsgRevokeFeeAllowance, &weightMsgRevokeFeeAllowance, nil,
// 		func(_ *rand.Rand) {
// 			weightMsgRevokeFeeAllowance = simappparams.DefaultWeightRevokeFeeAllowance
// 		},
// 	)

// 	return simulation.WeightedOperations{
// 		simulation.NewWeightedOperation(
// 			weightMsgGrantFeeAllowance,
// 			SimulateMsgGrantFeeAllowance(ak, bk, k),
// 		),
// 		simulation.NewWeightedOperation(
// 			weightMsgRevokeFeeAllowance,
// 			SimulateMsgRevokeFeeAllowance(ak, bk, k),
// 		),
// 	}
// }

// func SimulateMsgGrantFeeAllowance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
// 	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
// 		granter, _ := simtypes.RandomAcc(r, accs)
// 		grantee, _ := simtypes.RandomAcc(r, accs)

// 		account := ak.GetAccount(ctx, granter.Address)

// 		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
// 		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
// 		if err != nil {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgGrantFeeAllowance, err.Error()), nil, err
// 		}

// 		spendableCoins = spendableCoins.Sub(fees)

// 		msg, err := types.NewMsgGrantFeeAllowance(&types.BasicFeeAllowance{
// 			SpendLimit: spendableCoins,
// 			Expiration: types.ExpiresAtTime(ctx.BlockTime().Add(30 * time.Hour)),
// 		}, granter.Address, grantee.Address)

// 		if err != nil {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgGrantFeeAllowance, err.Error()), nil, err
// 		}
// 		txGen := simappparams.MakeTestEncodingConfig().TxConfig
// 		tx, err := helpers.GenTx(
// 			txGen,
// 			[]sdk.Msg{msg},
// 			fees,
// 			helpers.DefaultGenTxGas,
// 			chainID,
// 			[]uint64{account.GetAccountNumber()},
// 			[]uint64{account.GetSequence()},
// 			granter.PrivKey,
// 		)

// 		if err != nil {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgGrantFeeAllowance, "unable to generate mock tx"), nil, err
// 		}

// 		_, _, err = app.Deliver(txGen.TxEncoder(), tx)

// 		if err != nil {
// 			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
// 		}
// 		return simtypes.NewOperationMsg(msg, true, ""), nil, err
// 	}
// }

// func SimulateMsgRevokeFeeAllowance(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
// 	return func(
// 		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
// 	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

// 		hasGrant := false
// 		var granterAddr sdk.AccAddress
// 		var granteeAddr sdk.AccAddress
// 		k.IterateAllFeeAllowances(ctx, func(grant types.FeeAllowanceGrant) bool {

// 			granter, err := sdk.AccAddressFromBech32(grant.Granter)
// 			if err != nil {
// 				panic(err)
// 			}
// 			grantee, err := sdk.AccAddressFromBech32(grant.Grantee)
// 			if err != nil {
// 				panic(err)
// 			}
// 			granterAddr = granter
// 			granteeAddr = grantee
// 			hasGrant = true
// 			return true
// 		})

// 		if !hasGrant {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgRevokeFeeAllowance, "no grants"), nil, nil
// 		}
// 		granter, ok := simtypes.FindAccount(accs, granterAddr)

// 		if !ok {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgRevokeFeeAllowance, "Account not found"), nil, nil
// 		}

// 		account := ak.GetAccount(ctx, granter.Address)
// 		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
// 		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
// 		if err != nil {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgRevokeFeeAllowance, err.Error()), nil, err
// 		}

// 		msg := types.NewMsgRevokeFeeAllowance(granterAddr, granteeAddr)

// 		txGen := simappparams.MakeTestEncodingConfig().TxConfig
// 		tx, err := helpers.GenTx(
// 			txGen,
// 			[]sdk.Msg{&msg},
// 			fees,
// 			helpers.DefaultGenTxGas,
// 			chainID,
// 			[]uint64{account.GetAccountNumber()},
// 			[]uint64{account.GetSequence()},
// 			granter.PrivKey,
// 		)

// 		if err != nil {
// 			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgRevokeFeeAllowance, err.Error()), nil, err
// 		}

// 		_, _, err = app.Deliver(txGen.TxEncoder(), tx)
// 		return simtypes.NewOperationMsg(&msg, true, ""), nil, err
// 	}
// }
