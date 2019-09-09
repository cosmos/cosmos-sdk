package simulation

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateValidator = "op_weight_msg_create_validator"
	OpWeightMsgEditValidator   = "op_weight_msg_edit_validator"
	OpWeightMsgDelegate        = "op_weight_msg_delegate"
	OpWeightMsgUndelegate      = "op_weight_msg_undelegate"
	OpWeightMsgBeginRedelegate = "op_weight_msg_begin_redelegate"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper, k keeper.Keeper) simulation.WeightedOperations {

	var (
		weightMsgCreateValidator int
		weightMsgEditValidator   int
		weightMsgDelegate        int
		weightMsgUndelegate      int
		weightMsgBeginRedelegate int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil,
		func(_ *rand.Rand) { weightMsgCreateValidator = 100 })

	appParams.GetOrGenerate(cdc, OpWeightMsgEditValidator, &weightMsgEditValidator, nil,
		func(_ *rand.Rand) { weightMsgEditValidator = 5 })

	appParams.GetOrGenerate(cdc, OpWeightMsgDelegate, &weightMsgDelegate, nil,
		func(_ *rand.Rand) { weightMsgDelegate = 100 })

	appParams.GetOrGenerate(cdc, OpWeightMsgUndelegate, &weightMsgUndelegate, nil,
		func(_ *rand.Rand) { weightMsgUndelegate = 100 })

	appParams.GetOrGenerate(cdc, OpWeightMsgBeginRedelegate, &weightMsgBeginRedelegate, nil,
		func(_ *rand.Rand) { weightMsgBeginRedelegate = 100 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgEditValidator,
			SimulateMsgEditValidator(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgDelegate,
			SimulateMsgDelegate(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgUndelegate,
			SimulateMsgUndelegate(ak, k),
		),
		simulation.NewWeigthedOperation(
			weightMsgBeginRedelegate,
			SimulateMsgBeginRedelegate(ak, k),
		),
	}
}

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		simAccount, _ := simulation.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, found := k.GetValidator(ctx, address)
		if found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		denom := k.GetParams(ctx).BondDenom
		amount := ak.GetAccount(ctx, simAccount.Address).GetCoins().AmountOf(denom)
		if !amount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		amount, err = simulation.RandPositiveInt(r, amount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		coins := account.SpendableCoins(ctx.BlockTime())

		var fees sdk.Coins
		coins, hasNeg := coins.SafeSub(sdk.Coins{selfDelegation})
		if !hasNeg {
			fees, err = simulation.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		description := types.NewDescription(
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
		)

		maxCommission := sdk.NewDecWithPrec(r.Int63n(1000), 3)
		commission := types.NewCommissionRates(
			simulation.RandomDecAmount(r, maxCommission),
			maxCommission,
			simulation.RandomDecAmount(r, maxCommission),
		)

		msg := types.NewMsgCreateValidator(address, simAccount.PubKey,
			selfDelegation, description, commission, sdk.OneInt())

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgEditValidator generates a MsgEditValidator with random values
func SimulateMsgEditValidator(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		if len(k.GetAllValidators(ctx)) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		val, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		address := val.GetOperator()

		newCommissionRate := simulation.RandomDecAmount(r, val.Commission.MaxRate)
		err = val.Commission.ValidateNewRate(newCommissionRate, ctx.BlockHeader().Time)
		if err != nil {
			// skip as the commission is invalid
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, found := simulation.FindAccount(accs, sdk.AccAddress(val.GetOperator()))
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		description := types.NewDescription(
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
		)

		msg := types.NewMsgEditValidator(address, description, &newCommissionRate, nil)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgDelegate(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		denom := k.GetParams(ctx).BondDenom
		if len(k.GetAllValidators(ctx)) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, _ := simulation.RandomAcc(r, accs)
		val, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		if val.InvalidExRate() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		amount := ak.GetAccount(ctx, simAccount.Address).GetCoins().AmountOf(denom)
		if !amount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		amount, err = simulation.RandPositiveInt(r, amount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		bondAmt := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		coins := account.SpendableCoins(ctx.BlockTime())

		var fees sdk.Coins
		coins, hasNeg := coins.SafeSub(sdk.Coins{bondAmt})
		if !hasNeg {
			fees, err = simulation.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		msg := types.NewMsgDelegate(simAccount.Address, val.GetOperator(), bondAmt)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgUndelegate generates a MsgUndelegate with random values
func SimulateMsgUndelegate(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		simAccount, idx := simulation.RandomAcc(r, accs)
		delegations := k.GetAllDelegatorDelegations(ctx, simAccount.Address)

		var accsCopy []simulation.Account
		accsCopy = append(accsCopy, accs...)
		accsCopy = append(accsCopy[:idx], accsCopy[idx+1:]...)
		for len(accsCopy) > 0 && len(delegations) == 0 {
			simAccount, idx = simulation.RandomAcc(r, accsCopy)
			delegations = k.GetAllDelegatorDelegations(ctx, simAccount.Address)
			accsCopy = append(accsCopy[:idx], accsCopy[idx+1:]...)
		}

		if len(accsCopy) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		delegation := delegations[r.Intn(len(delegations))]

		if k.HasMaxUnbondingDelegationEntries(ctx, simAccount.Address, delegation.GetValidatorAddr()) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		validator, found := k.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", validator.GetOperator())
		}

		totalBond := validator.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		unbondAmt, err := simulation.RandPositiveInt(r, totalBond)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if unbondAmt.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgUndelegate(
			simAccount.Address, delegation.ValidatorAddress, sdk.NewCoin(k.BondDenom(ctx), unbondAmt),
		)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgBeginRedelegate generates a MsgBeginRedelegate with random values
func SimulateMsgBeginRedelegate(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		simAccount, idx := simulation.RandomAcc(r, accs)
		delegations := k.GetAllDelegatorDelegations(ctx, simAccount.Address)

		var accsCopy []simulation.Account
		accsCopy = append(accsCopy, accs...)
		accsCopy = append(accsCopy[:idx], accsCopy[idx+1:]...)
		for len(accsCopy) > 0 && len(delegations) == 0 {
			simAccount, idx = simulation.RandomAcc(r, accsCopy)
			delegations = k.GetAllDelegatorDelegations(ctx, simAccount.Address)
			accsCopy = append(accsCopy[:idx], accsCopy[idx+1:]...)
		}

		if len(accsCopy) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		delegation := delegations[r.Intn(len(delegations))]

		srcVal, found := k.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", srcVal.GetOperator())
		}

		if k.HasReceivingRedelegation(ctx, simAccount.Address, srcVal.GetOperator()) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil // skip
		}

		destVal, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		if srcVal.GetOperator().Equals(destVal.GetOperator()) ||
			destVal.InvalidExRate() ||
			k.HasMaxRedelegationEntries(ctx, simAccount.Address, srcVal.GetOperator(), destVal.GetOperator()) {

			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		redAmt, err := simulation.RandPositiveInt(r, totalBond)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if redAmt.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		// check if the shares truncate to zero
		shares, err := srcVal.SharesFromTokens(redAmt)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if srcVal.TokensFromShares(shares).TruncateInt().IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil // skip
		}

		// get tx fees
		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgBeginRedelegate(
			simAccount.Address, srcVal.GetOperator(), destVal.GetOperator(),
			sdk.NewCoin(k.BondDenom(ctx), redAmt),
		)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			[]crypto.PrivKey{simAccount.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
