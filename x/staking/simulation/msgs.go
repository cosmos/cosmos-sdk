package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	simutil "github.com/cosmos/cosmos-sdk/x/simulation/util"
)

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(m auth.AccountKeeper, k staking.Keeper) simulation.Operation {
	handler := staking.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		denom := k.GetParams(ctx).BondDenom
		description := staking.Description{
			Moniker: simutil.RandStringOfLength(r, 10),
		}

		maxCommission := sdk.NewDecWithPrec(r.Int63n(1000), 3)
		commission := staking.NewCommissionRates(
			simutil.RandomDecAmount(r, maxCommission),
			maxCommission,
			simutil.RandomDecAmount(r, maxCommission),
		)

		acc := simulation.RandomAcc(r, accs)
		address := sdk.ValAddress(acc.Address)
		amount := m.GetAccount(ctx, acc.Address).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simutil.RandomAmount(r, amount)
		}

		if amount.Equal(sdk.ZeroInt()) {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}

		selfDelegation := sdk.NewCoin(denom, amount)
		msg := staking.NewMsgCreateValidator(address, acc.PubKey,
			selfDelegation, description, commission, sdk.OneInt())

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(staking.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}

		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}

// SimulateMsgEditValidator generates a MsgEditValidator with random values
func SimulateMsgEditValidator(k staking.Keeper) simulation.Operation {
	handler := staking.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		description := staking.Description{
			Moniker:  simutil.RandStringOfLength(r, 10),
			Identity: simutil.RandStringOfLength(r, 10),
			Website:  simutil.RandStringOfLength(r, 10),
			Details:  simutil.RandStringOfLength(r, 10),
		}

		if len(k.GetAllValidators(ctx)) == 0 {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}
		val := keeper.RandomValidator(r, k, ctx)
		address := val.GetOperator()
		newCommissionRate := simutil.RandomDecAmount(r, val.Commission.MaxRate)

		msg := staking.NewMsgEditValidator(address, description, &newCommissionRate, nil)

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(staking.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}
		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}
		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgDelegate(m auth.AccountKeeper, k staking.Keeper) simulation.Operation {
	handler := staking.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		denom := k.GetParams(ctx).BondDenom
		if len(k.GetAllValidators(ctx)) == 0 {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}
		val := keeper.RandomValidator(r, k, ctx)
		validatorAddress := val.GetOperator()
		delegatorAcc := simulation.RandomAcc(r, accs)
		delegatorAddress := delegatorAcc.Address
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simutil.RandomAmount(r, amount)
		}
		if amount.Equal(sdk.ZeroInt()) {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}

		msg := staking.NewMsgDelegate(
			delegatorAddress, validatorAddress, sdk.NewCoin(denom, amount))

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(staking.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}
		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}
		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}

// SimulateMsgUndelegate generates a MsgUndelegate with random values
func SimulateMsgUndelegate(m auth.AccountKeeper, k staking.Keeper) simulation.Operation {
	handler := staking.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		delegatorAcc := simulation.RandomAcc(r, accs)
		delegatorAddress := delegatorAcc.Address
		delegations := k.GetAllDelegatorDelegations(ctx, delegatorAddress)
		if len(delegations) == 0 {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}
		delegation := delegations[r.Intn(len(delegations))]

		validator, found := k.GetValidator(ctx, delegation.GetValidatorAddr())
		if !found {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}

		totalBond := validator.TokensFromShares(delegation.GetShares()).TruncateInt()
		unbondAmt := simutil.RandomAmount(r, totalBond)
		if unbondAmt.Equal(sdk.ZeroInt()) {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}

		msg := staking.NewMsgUndelegate(
			delegatorAddress, delegation.ValidatorAddress, sdk.NewCoin(k.GetParams(ctx).BondDenom, unbondAmt),
		)
		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(staking.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s, got error %v",
				msg.GetSignBytes(), msg.ValidateBasic())
		}

		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}

		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}

// SimulateMsgBeginRedelegate generates a MsgBeginRedelegate with random values
func SimulateMsgBeginRedelegate(m auth.AccountKeeper, k staking.Keeper) simulation.Operation {
	handler := staking.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		denom := k.GetParams(ctx).BondDenom
		if len(k.GetAllValidators(ctx)) == 0 {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}
		srcVal := keeper.RandomValidator(r, k, ctx)
		srcValidatorAddress := srcVal.GetOperator()
		destVal := keeper.RandomValidator(r, k, ctx)
		destValidatorAddress := destVal.GetOperator()
		delegatorAcc := simulation.RandomAcc(r, accs)
		delegatorAddress := delegatorAcc.Address
		// TODO
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simutil.RandomAmount(r, amount)
		}
		if amount.Equal(sdk.ZeroInt()) {
			return simulation.NoOpMsg(staking.ModuleName), nil, nil
		}

		msg := staking.NewMsgBeginRedelegate(
			delegatorAddress, srcValidatorAddress, destValidatorAddress, sdk.NewCoin(denom, amount),
		)
		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(staking.ModuleName), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		}

		ctx, write := ctx.CacheContext()
		ok := handler(ctx, msg).IsOK()
		if ok {
			write()
		}

		opMsg = simulation.NewOperationMsg(msg, ok, "")
		return opMsg, nil, nil
	}
}
