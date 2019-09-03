package simulation

import (
	"errors"
	"math/rand"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// Simulation parameter constants
const (
	OpWeightMsgUnjail = "op_weight_msg_unjail"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	k keeper.Keeper, sk stakingkeeper.Keeper) simulation.WeightedOperations {

	var weightMsgUnjail int
	appParams.GetOrGenerate(cdc, OpWeightMsgUnjail, &weightMsgUnjail, nil,
		func(_ *rand.Rand) { weightMsgUnjail = 100 })

	return simulation.WeightedOperations{
		simulation.NewWeigthedOperation(
			weightMsgUnjail,
			SimulateMsgUnjail(ak, k, sk),
		),
	}
}

// SimulateMsgUnjail generates a MsgUnjail with random values
func SimulateMsgUnjail(ak types.AccountKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account,
		chainID string) (opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {
		// TODO: create iterator to get all jailed validators and then select a random
		// from the set
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

		msg := types.NewMsgUnjail(validator.GetOperator())

		account := ak.GetAccount(ctx, sdk.AccAddress(validator.GetOperator()))
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
