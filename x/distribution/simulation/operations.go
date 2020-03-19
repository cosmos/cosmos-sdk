package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"

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
	appParams module.AppParams, cdc *codec.Codec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper, sk stakingkeeper.Keeper,
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
			SimulateMsgSetWithdrawAddress(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawDelegationReward,
			SimulateMsgWithdrawDelegatorReward(ak, bk, k, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgWithdrawValidatorCommission,
			SimulateMsgWithdrawValidatorCommission(ak, bk, k, sk),
		),
		simulation.NewWeightedOperation(
			weightMsgFundCommunityPool,
			SimulateMsgFundCommunityPool(ak, bk, k, sk),
		),
	}
}

// SimulateMsgSetWithdrawAddress generates a MsgSetWithdrawAddress with random values.
func SimulateMsgSetWithdrawAddress(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) module.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []module.Account, chainID string,
	) (module.OperationMsg, []module.FutureOperation, error) {
		if !k.GetWithdrawAddrEnabled(ctx) {
			return module.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, _ := module.RandomAcc(r, accs)
		simToAccount, _ := module.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := module.RandomFees(r, ctx, spendable)
		if err != nil {
			return module.NoOpMsg(types.ModuleName), nil, err
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
			return module.NoOpMsg(types.ModuleName), nil, err
		}

		return module.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawDelegatorReward generates a MsgWithdrawDelegatorReward with random values.
func SimulateMsgWithdrawDelegatorReward(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) module.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []module.Account, chainID string,
	) (module.OperationMsg, []module.FutureOperation, error) {
		simAccount, _ := module.RandomAcc(r, accs)
		delegations := sk.GetAllDelegatorDelegations(ctx, simAccount.Address)
		if len(delegations) == 0 {
			return module.NoOpMsg(types.ModuleName), nil, nil
		}

		delegation := delegations[r.Intn(len(delegations))]

		validator := sk.Validator(ctx, delegation.GetValidatorAddr())
		if validator == nil {
			return module.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", delegation.GetValidatorAddr())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := module.RandomFees(r, ctx, spendable)
		if err != nil {
			return module.NoOpMsg(types.ModuleName), nil, err
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
			return module.NoOpMsg(types.ModuleName), nil, err
		}

		return module.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgWithdrawValidatorCommission generates a MsgWithdrawValidatorCommission with random values.
func SimulateMsgWithdrawValidatorCommission(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) module.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []module.Account, chainID string,
	) (module.OperationMsg, []module.FutureOperation, error) {

		validator, ok := stakingkeeper.RandomValidator(r, sk, ctx)
		if !ok {
			return module.NoOpMsg(types.ModuleName), nil, nil
		}

		commission := k.GetValidatorAccumulatedCommission(ctx, validator.GetOperator())
		if commission.Commission.IsZero() {
			return module.NoOpMsg(types.ModuleName), nil, nil
		}

		simAccount, found := module.FindAccount(accs, sdk.AccAddress(validator.GetOperator()))
		if !found {
			return module.NoOpMsg(types.ModuleName), nil, fmt.Errorf("validator %s not found", validator.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fees, err := module.RandomFees(r, ctx, spendable)
		if err != nil {
			return module.NoOpMsg(types.ModuleName), nil, err
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
			return module.NoOpMsg(types.ModuleName), nil, err
		}

		return module.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgFundCommunityPool simulates MsgFundCommunityPool execution where
// a random account sends a random amount of its funds to the community pool.
func SimulateMsgFundCommunityPool(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, sk stakingkeeper.Keeper) module.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []module.Account, chainID string,
	) (module.OperationMsg, []module.FutureOperation, error) {

		funder, _ := module.RandomAcc(r, accs)

		account := ak.GetAccount(ctx, funder.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		fundAmount := module.RandSubsetCoins(r, spendable)
		if fundAmount.Empty() {
			return module.NoOpMsg(types.ModuleName), nil, nil
		}

		var (
			fees sdk.Coins
			err  error
		)

		coins, hasNeg := spendable.SafeSub(fundAmount)
		if !hasNeg {
			fees, err = module.RandomFees(r, ctx, coins)
			if err != nil {
				return module.NoOpMsg(types.ModuleName), nil, err
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
			return module.NoOpMsg(types.ModuleName), nil, err
		}

		return module.NewOperationMsg(msg, true, ""), nil, nil
	}
}
