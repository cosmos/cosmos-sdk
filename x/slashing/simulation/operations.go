package simulation

import (
	"errors"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

// SimulateMsgUnjail generates a MsgUnjail with random values
func SimulateMsgUnjail(ak types.AccountKeeper, sk types.StakingKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {
		// TODO: create iterator to get all jailed validators and then select a random
		// from the set
		acc := simulation.RandomAcc(r, accs)
		validator := sk.Validator(ctx, sdk.ValAddress(acc.Address))
		if validator == nil {
			// skip as account is not a validator
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		if !validator.IsJailed() {
			// skip as validator is not jailed
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		msg := types.NewMsgUnjail(validator.GetOperator())

		fromAcc := ak.GetAccount(ctx, sdk.AccAddress(validator.GetOperator()))
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
