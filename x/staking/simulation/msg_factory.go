package simulation

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/types"
	"github.com/cosmos/cosmos-sdk/simsx"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgCreateValidatorFactory(k *keeper.Keeper) simsx.SimMsgFactoryFn[*types.MsgCreateValidator] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		r := testData.Rand()
		withoutValidators := simsx.SimAccountFilterFn(func(a simsx.SimAccount) bool {
			_, err := k.GetValidator(ctx, sdk.ValAddress(a.Address))
			return err != nil
		})
		withoutConsAddrUsed := simsx.SimAccountFilterFn(func(a simsx.SimAccount) bool {
			consPubKey := sdk.GetConsAddress(a.ConsKey.PubKey())
			_, err := k.GetValidatorByConsAddr(ctx, consPubKey)
			return err != nil
		})
		bondDenom, err := k.BondDenom(ctx)
		if err != nil {
			reporter.Skip("unable to determine bond denomination")
			return nil, nil
		}
		simAccount := testData.AnyAccount(reporter, withoutValidators, withoutConsAddrUsed, simsx.WithDenomBalance(bondDenom))
		if reporter.IsSkipped() {
			return nil, nil
		}
		selfDelegation := simAccount.LiquidBalance().RandSubsetCoin(reporter, bondDenom)
		description := types.NewDescription(
			r.StringN(10),
			r.StringN(10),
			r.StringN(10),
			r.StringN(10),
			r.StringN(10),
		)

		maxCommission := math.LegacyNewDecWithPrec(int64(r.IntInRange(0, 100)), 2)
		commission := types.NewCommissionRates(
			r.DecN(maxCommission),
			maxCommission,
			r.DecN(maxCommission),
		)

		addr, err := k.ValidatorAddressCodec().BytesToString(simAccount.Address)
		if err != nil {
			reporter.Skip("unable to generate validator address")
			return nil, nil
		}
		msg, err := types.NewMsgCreateValidator(addr, simAccount.ConsKey.PubKey(), selfDelegation, description, commission, math.OneInt())
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		return []simsx.SimAccount{simAccount}, msg
	}
}

func MsgDelegateFactory(k *keeper.Keeper) simsx.SimMsgFactoryFn[*types.MsgDelegate] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		r := testData.Rand()
		bondDenom, err := k.BondDenom(ctx)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		val := randomValidator(reporter, ctx, k, r)
		if reporter.IsSkipped() {
			return nil, nil
		}

		if !val.InvalidExRate() {
			reporter.Skip("validator's invalid exchange rate")
			return nil, nil
		}
		sender := testData.AnyAccount(reporter)
		delegation := sender.LiquidBalance().RandSubsetCoin(reporter, bondDenom)
		return []simsx.SimAccount{sender}, types.NewMsgDelegate(sender.AddressBech32, val.GetOperator(), delegation)
	}
}

func MsgUndelegateFactory(k *keeper.Keeper) simsx.SimMsgFactoryFn[*types.MsgUndelegate] {
	return func(ctx context.Context, testData *simsx.ChainDataSource, reporter simsx.SimulationReporter) ([]simsx.SimAccount, sdk.Msg) {
		r := testData.Rand()
		bondDenom, err := k.BondDenom(ctx)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		val := randomValidator(reporter, ctx, k, r)
		if reporter.IsSkipped() {
			return nil, nil
		}

		// select delegator and amount for undelegate
		valAddr, err := k.ValidatorAddressCodec().StringToBytes(val.GetOperator())
		if err != nil {
			reporter.Skipf("unable to get validator address: %s", err.Error())
			return nil, nil
		}
		delegations, err := k.GetValidatorDelegations(ctx, valAddr)
		if err != nil {
			reporter.Skipf("unable to get validator delegations: %s", err.Error())
			return nil, nil
		}
		if delegations == nil {
			reporter.Skip("no delegation entries")
			return nil, nil
		}
		// get random delegator from validator
		delegation := delegations[r.Intn(len(delegations))]
		delAddr := delegation.GetDelegatorAddr()
		delegator := testData.GetAccount(reporter, delAddr)

		hasMaxUD, err := k.HasMaxUnbondingDelegationEntries(ctx, delegator.Address, valAddr)
		if err != nil || hasMaxUD {
			reporter.Skipf("max unbodings or error fetching it: %s", err.Error())
			return nil, nil
		}

		totalBond := val.TokensFromShares(delegation.GetShares()).TruncateInt()
		if !totalBond.IsPositive() {
			reporter.Skip("total bond is negative")
			return nil, nil
		}

		unbondAmt, err := r.PositiveInt(totalBond)
		if err != nil {
			reporter.Skip("invalid unbond amount")
			return nil, nil
		}

		msg := types.NewMsgUndelegate(delAddr, val.GetOperator(), sdk.NewCoin(bondDenom, unbondAmt))
		return []simsx.SimAccount{delegator}, msg
	}
}

func randomValidator(reporter simsx.SimulationReporter, ctx context.Context, k *keeper.Keeper, r *simsx.XRand) types.Validator {
	vals, err := k.GetAllValidators(ctx)
	if err != nil || len(vals) == 0 {
		reporter.Skipf("unable to get validators or empty list: %s", err)
		return types.Validator{}
	}
	val, ok := testutil.RandSliceElem(r.Rand, vals)
	if !ok {
		reporter.Skip("validator is not ok")
		return types.Validator{}
	}
	return val
}
