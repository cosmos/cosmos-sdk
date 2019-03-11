package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulateMsgSetWithdrawAddress
func SimulateMsgSetWithdrawAddress(m auth.AccountKeeper, k distribution.Keeper) simulation.Operation {
	handler := distribution.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		accountOrigin := simulation.RandomAcc(r, accs)
		accountDestination := simulation.RandomAcc(r, accs)
		msg := distribution.NewMsgSetWithdrawAddress(accountOrigin.Address, accountDestination.Address)

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
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

// SimulateMsgWithdrawDelegatorReward
func SimulateMsgWithdrawDelegatorReward(m auth.AccountKeeper, k distribution.Keeper) simulation.Operation {
	handler := distribution.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		delegatorAccount := simulation.RandomAcc(r, accs)
		validatorAccount := simulation.RandomAcc(r, accs)
		msg := distribution.NewMsgWithdrawDelegatorReward(delegatorAccount.Address, sdk.ValAddress(validatorAccount.Address))

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
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

// SimulateMsgWithdrawValidatorCommission
func SimulateMsgWithdrawValidatorCommission(m auth.AccountKeeper, k distribution.Keeper) simulation.Operation {
	handler := distribution.NewHandler(k)
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		account := simulation.RandomAcc(r, accs)
		msg := distribution.NewMsgWithdrawValidatorCommission(sdk.ValAddress(account.Address))

		if msg.ValidateBasic() != nil {
			return simulation.NoOpMsg(), nil, fmt.Errorf("expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
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
