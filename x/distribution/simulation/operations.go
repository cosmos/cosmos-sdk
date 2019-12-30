package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
)

// Simulation operation weights constants
const (
	OpWeightMsgSetWithdrawAddress          = "op_weight_msg_set_withdraw_address"
	OpWeightMsgWithdrawDelegationReward    = "op_weight_msg_withdraw_delegation_reward"
	OpWeightMsgWithdrawValidatorCommission = "op_weight_msg_withdraw_validator_commission"
	OpWeightMsgFundCommunityPool           = "op_weight_msg_fund_community_pool"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simulation.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	k keeper.Keeper, sk stakingkeeper.Keeper,
) simulation.WeightedOperations {

	var weightMsgSetWithdrawAddress int
	appParams.GetOrGenerate(cdc, OpWeightMsgSetWithdrawAddress, &weightMsgSetWithdrawAddress, nil,
		func(_ *rand.Rand) {
			weightMsgSetWithdrawAddress = simappparams.DefaultWeightMsgSetWithdrawAddress
		},
	)

	var weightMsgWithdrawDelegationReward int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawDelegationReward, &weightMsgWithdrawDelegationReward, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawDelegationReward = simappparams.DefaultWeightMsgWithdrawDelegationReward
		},
	)

	var weightMsgWithdrawValidatorCommission int
	appParams.GetOrGenerate(cdc, OpWeightMsgWithdrawValidatorCommission, &weightMsgWithdrawValidatorCommission, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawValidatorCommission = simappparams.DefaultWeightMsgWithdrawValidatorCommission
		},
	)

	var weightMsgFundCommunityPool int
	appParams.GetOrGenerate(cdc, OpWeightMsgFundCommunityPool, &weightMsgFundCommunityPool, nil,
		func(_ *rand.Rand) {
			weightMsgFundCommunityPool = simappparams.DefaultWeightMsgFundCommunityPool
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgSetWithdrawAddress,
			SimulateMsgSetWithdrawAddress(ak, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawDelegationReward,
			SimulateMsgWithdrawDelegatorReward(ak, k, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawValidatorCommission,
			SimulateMsgWithdrawValidatorCommission(ak, k, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgFundCommunityPool,
			SimulateMsgFundCommunityPool(ak, k, sk),
		),
	}
}

// SimulateMsgSetWithdrawAddress generates a MsgSetWithdrawAddress with random values.
// nolint: funlen
func SimulateMsgSetWithdrawAddress(ak types.AccountKeeper, k keeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		if !k.GetWithdrawAddrEnabled(ctx) {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, _ := simulation.RandomAcc(r, accs)
		simToAccount, _ := simulation.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)

		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgSetWithdrawAddress(simAccount.Address, simToAccount.Address)

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawDelegatorReward generates a MsgWithdrawDelegatorReward with random values.
// nolint: funlen
func SimulateMsgWithdrawDelegatorReward(ak types.AccountKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {
		simAccount, _ := simulation.RandomAcc(r, accs)
		delegations := sk.GetAllDelegatorDelegations(ctx, simAccount.Address)
		if len(delegations) == 0 {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		delegation := delegations[r.Intn(len(delegations))]

		validator := sk.Validator(ctx, delegation.GetValidatorAddr())
		if validator == nil {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", delegation.GetValidatorAddr())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgWithdrawDelegatorReward(simAccount.Address, validator.GetOperator())

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawValidatorCommission generates a MsgWithdrawValidatorCommission with random values.
// nolint: funlen
func SimulateMsgWithdrawValidatorCommission(ak types.AccountKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		validator, ok := stakingkeeper.RandomValidator(r, sk, ctx)
		if !ok {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		commission := k.GetValidatorAccumulatedCommission(ctx, validator.GetOperator())
		if commission.IsZero() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, found := simulation.FindAccount(accs, sdk.AccAddress(validator.GetOperator()))
		if !found {
			return simulation.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", validator.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		fees, err := simulation.RandomFees(r, ctx, account.SpendableCoins(ctx.BlockTime()))
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		msg := types.NewMsgWithdrawValidatorCommission(validator.GetOperator())

		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgFundCommunityPool simulates MsgFundCommunityPool execution where
// a random account sends a random amount of its funds to the community pool.
func SimulateMsgFundCommunityPool(ak types.AccountKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) simulation.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simulation.Account, chainID string,
	) (simulation.OperationMsg, []simulation.FutureOperation, error) {

		funder, _ := simulation.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, funder.Address)
		coins := account.SpendableCoins(ctx.BlockTime())

		fundAmount := simulation.RandSubsetCoins(r, coins)
		if fundAmount.Empty() {
			return simulation.NoOpMsg(types.ModuleName), nil, nil
		}

		var (
			fees sdk.Coins
			err  error
		)

		coins, hasNeg := coins.SafeSub(fundAmount)
		if !hasNeg {
			fees, err = simulation.RandomFees(r, ctx, coins)
			if err != nil {
				return simulation.NoOpMsg(types.ModuleName), nil, err
			}
		}

		msg := types.NewMsgFundCommunityPool(fundAmount, funder.Address)
		tx := helpers.GenTx(
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			funder.PrivKey,
		)

		_, _, err = app.Deliver(tx)
		if err != nil {
			return simulation.NoOpMsg(types.ModuleName), nil, err
		}

		return simulation.NewOperationMsg(msg, true, ""), nil, nil
	}
}
