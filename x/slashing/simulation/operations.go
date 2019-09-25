package simulation

import (
	"errors"
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
func SimulateMsgUnjail(ak types.AccountKeeper,
	k keeper.Keeper, sk stakingkeeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		validator, ok := stakingkeeper.RandomValidator(r, sk, ctx)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil // skip
		}

		simAccount, found := simulation.FindAccount(accs, sdk.AccAddress(validator.GetOperator()))
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil // skip
		}

		if !validator.IsJailed() {
			// skip as validator is not jailed
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		consAddr := sdk.ConsAddress(validator.GetConsPubKey().Address())
		info, found := k.GetValidatorSigningInfo(ctx, consAddr)
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, nil // skip
		}

		selfDel := sk.Delegation(ctx, simAccount.Address, validator.GetOperator())
		if selfDel == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, nil // skip
		}

		// skip if:
		// - validator cannot be unjailed due to tombstone
		// - validator is still in jailed period
		// - self delegation too low
		if info.Tombstoned ||
			ctx.BlockHeader().Time.Before(info.JailedUntil) ||
			validator.TokensFromShares(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		account := ak.GetAccount(ctx, sdk.AccAddress(validator.GetOperator()))
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgUnjail(validator.GetOperator())

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
