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
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// SimulateMsgUnjail generates a MsgUnjail with random values
func SimulateMsgUnjail(ak types.AccountKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {
		// TODO: create iterator to get all jailed validators and then select a random
		// from the set
		validator := stakingkeeper.RandomValidator(r, sk, ctx)
		acc, found := simulation.FindAccount(accs, sdk.AccAddress(validator.GetOperator()))
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", validator.GetOperator())
		}

		if !validator.IsJailed() {
			// skip as validator is not jailed
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		consAddr := sdk.ConsAddress(validator.GetConsPubKey().Address())
		info, found := k.GetValidatorSigningInfo(ctx, consAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("slashing info not found for validator %s", consAddr)
		}

		switch {
		case validator.IsJailed() && info.Tombstoned:
			// skip as validator cannot be unjailed due to tombstone
			return simulation.NoOpMsg(types.ModuleName), nil, nil

		case validator.IsJailed() && ctx.BlockHeader().Time.Before(info.JailedUntil):
			// skip as validator is still in jailed period
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
