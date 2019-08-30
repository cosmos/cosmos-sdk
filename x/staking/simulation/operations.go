package simulation

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		simAccount := simulation.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, found := k.GetValidator(ctx, address)
		if found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		denom := k.GetParams(ctx).BondDenom
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

		amount := ak.GetAccount(ctx, simAccount.Address).GetCoins().AmountOf(denom)
		if !amount.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		amount, err = simulation.RandPositiveInt(r, amount)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		msg := types.NewMsgCreateValidator(address, simAccount.PubKey,
			selfDelegation, description, commission, sdk.OneInt())

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, sdk.Coins{selfDelegation})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

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

		val := keeper.RandomValidator(r, k, ctx)
		address := val.GetOperator()

		newCommissionRate := simulation.RandomDecAmount(r, val.Commission.MaxRate)
		err = val.Commission.ValidateNewRate(newCommissionRate, ctx.BlockHeader().Time)
		if err != nil {
			// skip as the commission is invalid
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		description := types.NewDescription(
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
		)

		msg := types.NewMsgEditValidator(address, description, &newCommissionRate, nil)

		simAccount, found := simulation.FindAccount(accs, sdk.AccAddress(val.GetOperator()))
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

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

		simAccount := simulation.RandomAcc(r, accs)

		val := keeper.RandomValidator(r, k, ctx)
		for val.InvalidExRate() {
			val = keeper.RandomValidator(r, k, ctx)
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
		msg := types.NewMsgDelegate(simAccount.Address, val.GetOperator(), bondAmt)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, sdk.Coins{bondAmt})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

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

		simAccount := simulation.RandomAcc(r, accs)
		delegations := k.GetAllDelegatorDelegations(ctx, simAccount.Address)
		if len(delegations) == 0 {
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

		totalBond, err = simulation.RandPositiveInt(r, totalBond)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		unbondAmt, err := simulation.RandPositiveInt(r, totalBond)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if unbondAmt.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgUndelegate(
			simAccount.Address, delegation.ValidatorAddress, sdk.NewCoin(k.BondDenom(ctx), unbondAmt),
		)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

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

		simAccount := simulation.RandomAcc(r, accs)
		delegations := k.GetAllDelegatorDelegations(ctx, simAccount.Address)
		if len(delegations) == 0 {
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

		destVal := keeper.RandomValidator(r, k, ctx)
		for srcVal.GetOperator().Equals(destVal.GetOperator()) ||
			destVal.InvalidExRate() ||
			k.HasMaxRedelegationEntries(ctx, simAccount.Address, srcVal.GetOperator(), destVal.GetOperator()) {
			destVal = keeper.RandomValidator(r, k, ctx)
		}

		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		totalBond, err = simulation.RandPositiveInt(r, totalBond)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		redAmt, err := simulation.RandPositiveInt(r, totalBond)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		if redAmt.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgBeginRedelegate(
			simAccount.Address, srcVal.GetOperator(), destVal.GetOperator(),
			sdk.NewCoin(k.BondDenom(ctx), redAmt),
		)

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := helpers.RandomFees(r, ctx, account, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

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
