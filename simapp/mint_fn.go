package simapp

import (
	"context"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/math"
	authtypes "cosmossdk.io/x/auth/types"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"
	stakingtypes "cosmossdk.io/x/staking/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ProvideExampleMintFn returns the function used in x/mint's endblocker to mint new tokens.
// Note that this function can not have the mint keeper as a parameter because it would create a cyclic dependency.
func ProvideExampleMintFn(bankKeeper bankkeeper.Keeper) minttypes.MintFn {
	return func(ctx context.Context, env appmodule.Environment, minter *minttypes.Minter, epochId string, epochNumber int64) error {
		// in this example we ignore epochNumber as we don't care what epoch we are in, we just assume we are being called every minute.
		if epochId != "minute" {
			return nil
		}

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

		var mintParams minttypes.QueryParamsResponse
		err = env.RouterService.QueryRouterService().InvokeTyped(ctx, &minttypes.QueryParamsRequest{}, &mintParams)
		if err != nil {
			return err
		}

		var stakingPool stakingtypes.QueryPoolResponse
		err = env.RouterService.QueryRouterService().InvokeTyped(ctx, &stakingtypes.QueryPoolRequest{}, &stakingPool)
		if err != nil {
			return err
		}

		// bondedRatio
		bondedRatio := math.LegacyNewDecFromInt(stakingPool.Pool.BondedTokens).QuoInt(stakingTokenSupply.Amount)
		minter.Inflation = minter.NextInflationRate(mintParams.Params, bondedRatio)
		minter.AnnualProvisions = minter.NextAnnualProvisions(mintParams.Params, stakingTokenSupply.Amount)

		// because we are minting every minute, we need to divide the annual provisions by minutes in a year (525600)
		provisionAmt := minter.AnnualProvisions.QuoInt64(525600)
		mintedCoin := sdk.NewCoin(mintParams.Params.MintDenom, provisionAmt.TruncateInt())
		mintedCoins := sdk.NewCoins(mintedCoin)
		maxSupply := mintParams.Params.MaxSupply
		totalSupply := stakingTokenSupply.Amount

		// if maxSupply is not infinite, check against max_supply parameter
		if !maxSupply.IsZero() {
			if totalSupply.Add(mintedCoins.AmountOf(mintParams.Params.MintDenom)).GT(maxSupply) {
				// calculate the difference between maxSupply and totalSupply
				diff := maxSupply.Sub(totalSupply)
				// mint the difference
				diffCoin := sdk.NewCoin(mintParams.Params.MintDenom, diff)
				diffCoins := sdk.NewCoins(diffCoin)

				// mint coins
				if diffCoins.Empty() {
					// skip as no coins need to be minted
					return nil
				}

				if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, diffCoins); err != nil {
					return err
				}
				mintedCoins = diffCoins
			}
		}

		// mint coins if maxSupply is infinite or total staking supply is less than maxSupply
		if maxSupply.IsZero() || totalSupply.Add(mintedCoins.AmountOf(mintParams.Params.MintDenom)).LT(maxSupply) {
			// mint coins
			if mintedCoins.Empty() {
				// skip as no coins need to be minted
				return nil
			}

			if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, mintedCoins); err != nil {
				return err
			}
		}

		// Example of custom send while minting
		// Send some tokens to a "team account"
		// if err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, ... ); err != nil {
		// 	return err
		// }

		// TODO: figure how to get FeeCollectorName from mint module without generating a cyclic dependency
		if err = bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, mintedCoins); err != nil {
			return err
		}

		if mintedCoin.Amount.IsInt64() {
			defer telemetry.ModuleSetGauge(minttypes.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
		}

		return env.EventService.EventManager(ctx).EmitKV(
			minttypes.EventTypeMint,
			event.NewAttribute(minttypes.AttributeKeyBondedRatio, bondedRatio.String()),
			event.NewAttribute(minttypes.AttributeKeyInflation, minter.Inflation.String()),
			event.NewAttribute(minttypes.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			event.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		)
	}
}
