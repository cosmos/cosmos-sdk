package types

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InflationCalculationFn defines the function required to calculate inflation rate during
// BeginBlock. It receives the minter and params stored in the keeper, along with the current
// bondedRatio and returns the newly calculated inflation rate.
// It can be used to specify a custom inflation calculation logic, instead of relying on the
// default logic provided by the sdk.
type InflationCalculationFn func(ctx context.Context, minter Minter, params Params, bondedRatio math.LegacyDec) math.LegacyDec

type MintFn func(ctx context.Context, env appmodule.Environment, minter *Minter, params Params, mintCoins func(ctx context.Context, newCoins sdk.Coins) error, addCollectedFees func(context.Context, sdk.Coins) error) error

func DefaultMintFn(ctx context.Context, env appmodule.Environment, minter *Minter, params Params, mintCoins func(context.Context, sdk.Coins) error, addCollectedFees func(context.Context, sdk.Coins) error) error {
	var stakingParams stakingtypes.QueryParamsResponse
	err := env.RouterService.QueryRouterService().InvokeTyped(ctx, &stakingtypes.QueryParamsRequest{}, &stakingParams)
	if err != nil {
		return err
	}

	var bankSupply banktypes.QuerySupplyOfResponse
	err = env.RouterService.QueryRouterService().InvokeTyped(ctx, &banktypes.QuerySupplyOfRequest{Denom: stakingParams.Params.BondDenom}, &bankSupply)
	if err != nil {
		return err
	}
	stakingTokenSupply := bankSupply.Amount

	var stakingPool stakingtypes.QueryPoolResponse
	err = env.RouterService.QueryRouterService().InvokeTyped(ctx, &stakingtypes.QueryPoolRequest{}, &stakingPool)
	if err != nil {
		return err
	}

	// bondedRatio
	bondedRatio := math.LegacyNewDecFromInt(stakingPool.Pool.BondedTokens).QuoInt(stakingTokenSupply.Amount)
	minter.Inflation = minter.NextInflationRate(params, bondedRatio)
	minter.AnnualProvisions = minter.NextAnnualProvisions(params, stakingTokenSupply.Amount)
	// TODO: store minter afterwards

	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)
	maxSupply := params.MaxSupply
	totalSupply := stakingTokenSupply.Amount

	// if maxSupply is not infinite, check against max_supply parameter
	if !maxSupply.IsZero() {
		if totalSupply.Add(mintedCoins.AmountOf(params.MintDenom)).GT(maxSupply) {
			// calculate the difference between maxSupply and totalSupply
			diff := maxSupply.Sub(totalSupply)
			// mint the difference
			diffCoin := sdk.NewCoin(params.MintDenom, diff)
			diffCoins := sdk.NewCoins(diffCoin)

			// mint coins
			if err := mintCoins(ctx, diffCoins); err != nil {
				return err
			}
			mintedCoins = diffCoins
		}
	}

	// mint coins if maxSupply is infinite or total staking supply is less than maxSupply
	if maxSupply.IsZero() || totalSupply.Add(mintedCoins.AmountOf(params.MintDenom)).LT(maxSupply) {
		// mint coins
		if err := mintCoins(ctx, mintedCoins); err != nil {
			return err
		}
	}

	// send the minted coins to the fee collector account
	// TODO: figure out a better way to do this
	err = addCollectedFees(ctx, mintedCoins)
	if err != nil {
		return err
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	return env.EventService.EventManager(ctx).EmitKV(
		EventTypeMint,
		event.NewAttribute(AttributeKeyBondedRatio, bondedRatio.String()),
		event.NewAttribute(AttributeKeyInflation, minter.Inflation.String()),
		event.NewAttribute(AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
		event.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
	)
}

// DefaultInflationCalculationFn is the default function used to calculate inflation.
func DefaultInflationCalculationFn(_ context.Context, minter Minter, params Params, bondedRatio math.LegacyDec) math.LegacyDec {
	return minter.NextInflationRate(params, bondedRatio)
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(minter Minter, params Params) *GenesisState {
	return &GenesisState{
		Minter: minter,
		Params: params,
	}
}

// DefaultGenesisState creates a default GenesisState object
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Minter: DefaultInitialMinter(),
		Params: DefaultParams(),
	}
}

// ValidateGenesis validates the provided genesis state to ensure the
// expected invariants holds.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	return ValidateMinter(data.Minter)
}
