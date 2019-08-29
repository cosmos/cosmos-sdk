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

		acc := simulation.RandomAcc(r, accs)
		address := sdk.ValAddress(acc.Address)

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

		amount := ak.GetAccount(ctx, acc.Address).GetCoins().AmountOf(denom)
		switch {
		case amount.IsZero():
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case amount.IsPositive():
			amount = simulation.RandomAmount(r, amount)
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		msg := types.NewMsgCreateValidator(address, acc.PubKey,
			selfDelegation, description, commission, sdk.OneInt())

		fromAcc := ak.GetAccount(ctx, acc.Address)
		fees, err := helpers.RandomFees(r, ctx, fromAcc, sdk.Coins{selfDelegation})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{acc.PrivKey}...,
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

		description := types.NewDescription(
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
		)

		if len(k.GetAllValidators(ctx)) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		val := keeper.RandomValidator(r, k, ctx)
		address := val.GetOperator()
		newCommissionRate := simulation.RandomDecAmount(r, val.Commission.MaxRate)

		msg := types.NewMsgEditValidator(address, description, &newCommissionRate, nil)

		acc, found := simulation.FindAccount(accs, sdk.AccAddress(val.GetOperator()))
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		fromAcc := ak.GetAccount(ctx, acc.Address)
		fees, err := helpers.RandomFees(r, ctx, fromAcc, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{acc.PrivKey}...,
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

		val := keeper.RandomValidator(r, k, ctx)
		validatorAddress := val.GetOperator()
		delegatorAcc := simulation.RandomAcc(r, accs)
		delegatorAddress := delegatorAcc.Address

		amount := ak.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		switch {
		case amount.IsZero():
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case amount.IsPositive():
			amount = simulation.RandomAmount(r, amount)
		}

		bondAmt := sdk.NewCoin(denom, amount)
		msg := types.NewMsgDelegate(delegatorAddress, validatorAddress, bondAmt)

		fromAcc := ak.GetAccount(ctx, delegatorAcc.Address)
		fees, err := helpers.RandomFees(r, ctx, fromAcc, sdk.Coins{bondAmt})
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{delegatorAcc.PrivKey}...,
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

		delegatorAcc := simulation.RandomAcc(r, accs)
		delegations := k.GetAllDelegatorDelegations(ctx, delegatorAcc.Address)
		if len(delegations) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		delegation := delegations[r.Intn(len(delegations))]

		validator, found := k.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		totalBond := validator.TokensFromShares(delegation.GetShares()).TruncateInt()
		switch {
		case totalBond.IsZero():
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case totalBond.IsPositive():
			totalBond = simulation.RandomAmount(r, totalBond)
		}
		
		unbondAmt := simulation.RandomAmount(r, totalBond)
		if unbondAmt.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgUndelegate(
			delegatorAcc.Address, delegation.ValidatorAddress, sdk.NewCoin(k.BondDenom(ctx), unbondAmt),
		)

		fromAcc := ak.GetAccount(ctx, delegatorAcc.Address)
		fees, err := helpers.RandomFees(r, ctx, fromAcc, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{delegatorAcc.PrivKey}...,
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

		delegatorAcc := simulation.RandomAcc(r, accs)
		delegations := k.GetAllDelegatorDelegations(ctx, delegatorAcc.Address)
		if len(delegations) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		delegation := delegations[r.Intn(len(delegations))]

		srcVal, found := k.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		destVal := keeper.RandomValidator(r, k, ctx)

		for srcVal.GetOperator().Equals(destVal.GetOperator()) {
			destVal = keeper.RandomValidator(r, k, ctx)
		}

		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		switch {
		case totalBond.IsZero():
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		case totalBond.IsPositive():
			totalBond = simulation.RandomAmount(r, totalBond)
		}
		
		
		redAmt := simulation.RandomAmount(r, totalBond)
		if redAmt.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgBeginRedelegate(
			delegatorAcc.Address, srcVal.GetOperator(), destVal.GetOperator(),
			sdk.NewCoin(k.BondDenom(ctx), redAmt),
		)

		fromAcc := ak.GetAccount(ctx, delegatorAcc.Address)
		fees, err := helpers.RandomFees(r, ctx, fromAcc, nil)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			chainID,
			[]uint64{fromAcc.GetAccountNumber()},
			[]uint64{fromAcc.GetSequence()},
			[]crypto.PrivKey{delegatorAcc.PrivKey}...,
		)

		res := app.Deliver(tx)
		if !res.IsOK() {
			return simulation.NoOpMsg(types.ModuleName), nil, errors.New(res.Log)
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
