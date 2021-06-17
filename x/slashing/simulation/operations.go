package simulation

import (
	"errors"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// Simulation operation weights constants
const (
	OpWeightMsgUnjail = "op_weight_msg_unjail"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONMarshaler, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper, sk stakingkeeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgUnjail int
	appParams.GetOrGenerate(cdc, OpWeightMsgUnjail, &weightMsgUnjail, nil,
		func(_ *rand.Rand) {
			weightMsgUnjail = simappparams.DefaultWeightMsgUnjail
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgUnjail,
			SimulateMsgUnjail(ak, bk, k, sk),
		),
	}
}

// SimulateMsgUnjail generates a MsgUnjail with random values
func SimulateMsgUnjail(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

		validator, ok := stakingkeeper.RandomValidator(r, sk, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "validator is not ok"), nil, nil // skip
		}

		simAccount, found := simtypes.FindAccount(accs, sdk.AccAddress(validator.GetOperator()))
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "unable to find account"), nil, nil // skip
		}

		if !validator.IsJailed() {
			// TODO: due to this condition this message is almost, if not always, skipped !
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "validator is not jailed"), nil, nil
		}

		consAddr, err := validator.GetConsAddr()
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "unable to get validator consensus key"), nil, err
		}
		info, found := k.GetValidatorSigningInfo(ctx, consAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "unable to find validator signing info"), nil, nil // skip
		}

		selfDel := sk.Delegation(ctx, simAccount.Address, validator.GetOperator())
		if selfDel == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "self delegation is nil"), nil, nil // skip
		}

		account := ak.GetAccount(ctx, sdk.AccAddress(validator.GetOperator()))
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUnjail, "unable to generate fees"), nil, err
		}

		msg := types.NewMsgUnjail(validator.GetOperator())

		txGen := simappparams.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenTx(
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, res, err := app.Deliver(txGen.TxEncoder(), tx)

		// result should fail if:
		// - validator cannot be unjailed due to tombstone
		// - validator is still in jailed period
		// - self delegation too low
		if info.Tombstoned ||
			ctx.BlockHeader().Time.Before(info.JailedUntil) ||
			validator.TokensFromShares(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
			if res != nil && err == nil {
				if info.Tombstoned {
					return simtypes.NewOperationMsg(msg, true, ""), nil, errors.New("validator should not have been unjailed if validator tombstoned")
				}
				if ctx.BlockHeader().Time.Before(info.JailedUntil) {
					return simtypes.NewOperationMsg(msg, true, ""), nil, errors.New("validator unjailed while validator still in jail period")
				}
				if validator.TokensFromShares(selfDel.GetShares()).TruncateInt().LT(validator.GetMinSelfDelegation()) {
					return simtypes.NewOperationMsg(msg, true, ""), nil, errors.New("validator unjailed even though self-delegation too low")
				}
			}
			// msg failed as expected
			return simtypes.NewOperationMsg(msg, false, ""), nil, nil
		}

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, errors.New(res.Log)
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}
