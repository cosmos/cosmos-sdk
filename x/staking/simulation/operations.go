package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	DefaultWeightMsgTokenizeShares              int = 100
	DefaultWeightMsgRedeemTokensforShares       int = 100
	DefaultWeightMsgTransferTokenizeShareRecord int = 50
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateValidator             = "op_weight_msg_create_validator"
	OpWeightMsgEditValidator               = "op_weight_msg_edit_validator"
	OpWeightMsgDelegate                    = "op_weight_msg_delegate"
	OpWeightMsgUndelegate                  = "op_weight_msg_undelegate"
	OpWeightMsgBeginRedelegate             = "op_weight_msg_begin_redelegate"
	OpWeightMsgTokenizeShares              = "op_weight_msg_tokenize_shares"
	OpWeightMsgRedeemTokensforShares       = "op_weight_msg_redeem_tokens_for_shares"
	OpWeightMsgTransferTokenizeShareRecord = "op_weight_msg_transfer_tokenize_share_record"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams, cdc codec.JSONCodec, ak types.AccountKeeper,
	bk types.BankKeeper, k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateValidator             int
		weightMsgEditValidator               int
		weightMsgDelegate                    int
		weightMsgUndelegate                  int
		weightMsgBeginRedelegate             int
		weightMsgTokenizeShares              int
		weightMsgRedeemTokensforShares       int
		weightMsgTransferTokenizeShareRecord int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateValidator, &weightMsgCreateValidator, nil,
		func(_ *rand.Rand) {
			weightMsgCreateValidator = simappparams.DefaultWeightMsgCreateValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgEditValidator, &weightMsgEditValidator, nil,
		func(_ *rand.Rand) {
			weightMsgEditValidator = simappparams.DefaultWeightMsgEditValidator
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgDelegate, &weightMsgDelegate, nil,
		func(_ *rand.Rand) {
			weightMsgDelegate = simappparams.DefaultWeightMsgDelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgUndelegate, &weightMsgUndelegate, nil,
		func(_ *rand.Rand) {
			weightMsgUndelegate = simappparams.DefaultWeightMsgUndelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgBeginRedelegate, &weightMsgBeginRedelegate, nil,
		func(_ *rand.Rand) {
			weightMsgBeginRedelegate = simappparams.DefaultWeightMsgBeginRedelegate
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgTokenizeShares, &weightMsgTokenizeShares, nil,
		func(_ *rand.Rand) {
			weightMsgTokenizeShares = DefaultWeightMsgTokenizeShares
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgRedeemTokensforShares, &weightMsgRedeemTokensforShares, nil,
		func(_ *rand.Rand) {
			weightMsgRedeemTokensforShares = DefaultWeightMsgRedeemTokensforShares
		},
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgTransferTokenizeShareRecord, &weightMsgTransferTokenizeShareRecord, nil,
		func(_ *rand.Rand) {
			weightMsgTransferTokenizeShareRecord = DefaultWeightMsgTransferTokenizeShareRecord
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateValidator,
			SimulateMsgCreateValidator(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgEditValidator,
			SimulateMsgEditValidator(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDelegate,
			SimulateMsgDelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUndelegate,
			SimulateMsgUndelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBeginRedelegate,
			SimulateMsgBeginRedelegate(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgTokenizeShares,
			SimulateMsgTokenizeShares(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgRedeemTokensforShares,
			SimulateMsgRedeemTokensforShares(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgTransferTokenizeShareRecord,
			SimulateMsgTransferTokenizeShareRecord(ak, bk, k),
		),
	}
}

// SimulateMsgCreateValidator generates a MsgCreateValidator with random values
func SimulateMsgCreateValidator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator doesn't exist already
		_, found := k.GetValidator(ctx, address)
		if found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to find validator"), nil, nil
		}

		denom := k.GetParams(ctx).BondDenom

		balance := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !balance.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "balance is negative"), nil, nil
		}

		amount, err := simtypes.RandPositiveInt(r, balance)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate positive amount"), nil, err
		}

		selfDelegation := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		coins, hasNeg := spendable.SafeSub(sdk.Coins{selfDelegation})
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateValidator, "unable to generate fees"), nil, err
			}
		}

		description := types.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		maxCommission := sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 0, 100)), 2)
		commission := types.NewCommissionRates(
			simtypes.RandomDecAmount(r, maxCommission),
			maxCommission,
			simtypes.RandomDecAmount(r, maxCommission),
		)

		msg, err := types.NewMsgCreateValidator(address, simAccount.ConsKey.PubKey(), selfDelegation, description, commission, sdk.OneInt())
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to create CreateValidator message"), nil, err
		}

		txCtx := simulation.OperationInput{
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgEditValidator generates a MsgEditValidator with random values
func SimulateMsgEditValidator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		if len(k.GetAllValidators(ctx)) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "number of validators equal zero"), nil, nil
		}

		val, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "unable to pick a validator"), nil, nil
		}

		address := val.GetOperator()

		newCommissionRate := simtypes.RandomDecAmount(r, val.Commission.MaxRate)

		if err := val.Commission.ValidateNewRate(newCommissionRate, ctx.BlockHeader().Time); err != nil {
			// skip as the commission is invalid
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "invalid commission rate"), nil, nil
		}

		simAccount, found := simtypes.FindAccount(accs, sdk.AccAddress(val.GetOperator()))
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgEditValidator, "unable to find account"), nil, fmt.Errorf("validator %s not found", val.GetOperator())
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		description := types.NewDescription(
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
			simtypes.RandStringOfLength(r, 10),
		)

		msg := types.NewMsgEditValidator(address, description, &newCommissionRate, nil)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDelegate generates a MsgDelegate with random values
func SimulateMsgDelegate(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		denom := k.GetParams(ctx).BondDenom

		if len(k.GetAllValidators(ctx)) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "number of validators equal zero"), nil, nil
		}

		simAccount, _ := simtypes.RandomAcc(r, accs)
		val, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to pick a validator"), nil, nil
		}

		if val.InvalidExRate() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "validator's invalid echange rate"), nil, nil
		}

		amount := bk.GetBalance(ctx, simAccount.Address, denom).Amount
		if !amount.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "balance is negative"), nil, nil
		}

		amount, err := simtypes.RandPositiveInt(r, amount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to generate positive amount"), nil, err
		}

		bondAmt := sdk.NewCoin(denom, amount)

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		var fees sdk.Coins

		coins, hasNeg := spendable.SafeSub(sdk.Coins{bondAmt})
		if !hasNeg {
			fees, err = simtypes.RandomFees(r, ctx, coins)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegate, "unable to generate fees"), nil, err
			}
		}

		msg := types.NewMsgDelegate(simAccount.Address, val.GetOperator(), bondAmt)

		txCtx := simulation.OperationInput{
			App:           app,
			TxGen:         simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTx(txCtx, fees)
	}
}

// SimulateMsgUndelegate generates a MsgUndelegate with random values
func SimulateMsgUndelegate(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// get random validator
		validator, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "validator is not ok"), nil, nil
		}

		valAddr := validator.GetOperator()
		delegations := k.GetValidatorDelegations(ctx, validator.GetOperator())
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "keeper does have any delegation entries"), nil, nil
		}

		// get random delegator from validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		if k.HasMaxUnbondingDelegationEntries(ctx, delAddr, valAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "keeper does have a max unbonding delegation entries"), nil, nil
		}

		totalBond := validator.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "total bond is negative"), nil, nil
		}

		unbondAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "invalid unbond amount"), nil, err
		}

		if unbondAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUndelegate, "unbond amount is zero"), nil, nil
		}

		msg := types.NewMsgUndelegate(
			delAddr, valAddr, sdk.NewCoin(k.BondDenom(ctx), unbondAmt),
		)

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}
		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "account private key is nil"), nil, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		account := ak.GetAccount(ctx, delAddr)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBeginRedelegate generates a MsgBeginRedelegate with random values
func SimulateMsgBeginRedelegate(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// get random source validator
		srcVal, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "unable to pick validator"), nil, nil
		}

		srcAddr := srcVal.GetOperator()
		delegations := k.GetValidatorDelegations(ctx, srcAddr)
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "keeper does have any delegation entries"), nil, nil
		}

		// get random delegator from src validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		if k.HasReceivingRedelegation(ctx, delAddr, srcAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "receveing redelegation is not allowed"), nil, nil // skip
		}

		// get random destination validator
		destVal, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "unable to pick validator"), nil, nil
		}

		destAddr := destVal.GetOperator()
		if srcAddr.Equals(destAddr) || destVal.InvalidExRate() || k.HasMaxRedelegationEntries(ctx, delAddr, srcAddr, destAddr) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "checks failed"), nil, nil
		}

		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "total bond is negative"), nil, nil
		}

		redAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "unable to generate positive amount"), nil, err
		}

		if redAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "amount is zero"), nil, nil
		}

		// check if the shares truncate to zero
		shares, err := srcVal.SharesFromTokens(redAmt)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "invalid shares"), nil, err
		}

		if srcVal.TokensFromShares(shares).TruncateInt().IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "shares truncate to zero"), nil, nil // skip
		}

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}

		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBeginRedelegate, "account private key is nil"), nil, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		account := ak.GetAccount(ctx, delAddr)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := types.NewMsgBeginRedelegate(
			delAddr, srcAddr, destAddr,
			sdk.NewCoin(k.BondDenom(ctx), redAmt),
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgTokenizeShares generates a MsgTokenizeShares with random values
func SimulateMsgTokenizeShares(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// get random source validator
		srcVal, ok := keeper.RandomValidator(r, k, ctx)
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "unable to pick validator"), nil, nil
		}

		srcAddr := srcVal.GetOperator()
		delegations := k.GetValidatorDelegations(ctx, srcAddr)
		if delegations == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "keeper does have any delegation entries"), nil, nil
		}

		// get random delegator from src validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()

		// get random destination validator
		totalBond := srcVal.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "total bond is negative"), nil, nil
		}

		tokenizeShareAmt, err := simtypes.RandPositiveInt(r, totalBond)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "unable to generate positive amount"), nil, err
		}

		if tokenizeShareAmt.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "amount is zero"), nil, nil
		}

		// check if the shares truncate to zero
		shares, err := srcVal.SharesFromTokens(tokenizeShareAmt)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "invalid shares"), nil, err
		}

		if srcVal.TokensFromShares(shares).TruncateInt().IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "shares truncate to zero"), nil, nil // skip
		}

		// need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delAddr) {
				simAccount = simAcc
				break
			}
		}

		// if simaccount.PrivKey == nil, delegation address does not exist in accs. Return error
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTokenizeShares, "account private key is nil"), nil, fmt.Errorf("delegation addr: %s does not exist in simulation accounts", delAddr)
		}

		msg := &types.MsgTokenizeShares{
			DelegatorAddress:    delAddr.String(),
			ValidatorAddress:    srcAddr.String(),
			Amount:              sdk.NewCoin(k.BondDenom(ctx), tokenizeShareAmt),
			TokenizedShareOwner: delAddr.String(),
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgRedeemTokensforShares generates a MsgRedeemTokensforShares with random values
func SimulateMsgRedeemTokensforShares(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		redeemUser := simtypes.Account{}
		redeemCoin := sdk.Coin{}

		records := k.GetAllTokenizeShareRecords(ctx)
		if len(records) > 0 {
			record := records[r.Intn(len(records))]
			for _, acc := range accs {
				balance := bk.GetBalance(ctx, acc.Address, record.ShareTokenDenom)
				if balance.Amount.IsPositive() {
					redeemUser = acc
					redeemAmount, err := simtypes.RandPositiveInt(r, balance.Amount)
					if err == nil {
						redeemCoin = sdk.NewCoin(record.ShareTokenDenom, redeemAmount)
					}
					break
				}
			}
		}

		// if redeemUser.PrivKey == nil, redeem user does not exist in accs
		if redeemUser.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgRedeemTokensforShares, "account private key is nil"), nil, nil
		}

		if redeemCoin.Amount.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgRedeemTokensforShares, "empty balance in tokens"), nil, nil
		}

		account := ak.GetAccount(ctx, redeemUser.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := &types.MsgRedeemTokensforShares{
			DelegatorAddress: redeemUser.Address.String(),
			Amount:           redeemCoin,
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      redeemUser,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgTransferTokenizeShareRecord generates a MsgTransferTokenizeShareRecord with random values
func SimulateMsgTransferTokenizeShareRecord(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		destAccount, _ := simtypes.RandomAcc(r, accs)
		transferRecord := types.TokenizeShareRecord{}

		records := k.GetAllTokenizeShareRecords(ctx)
		if len(records) > 0 {
			record := records[r.Intn(len(records))]
			for _, acc := range accs {
				balance := bk.GetBalance(ctx, acc.Address, record.ShareTokenDenom)
				if balance.Amount.IsPositive() {
					simAccount = acc
					transferRecord = record
					break
				}
			}
		}

		// if simAccount.PrivKey == nil, record owner does not exist in accs
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTransferTokenizeShareRecord, "account private key is nil"), nil, nil
		}

		if transferRecord.Id == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTransferTokenizeShareRecord, "share record not found"), nil, nil
		}

		account := ak.GetAccount(ctx, simAccount.Address)
		spendable := bk.SpendableCoins(ctx, account.GetAddress())

		msg := &types.MsgTransferTokenizeShareRecord{
			TokenizeShareRecordId: transferRecord.Id,
			Sender:                simAccount.Address.String(),
			NewOwner:              destAccount.Address.String(),
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: spendable,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
