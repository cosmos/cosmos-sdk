package simapp

import (
	"context"
	"encoding/binary"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	minttypes "cosmossdk.io/x/mint/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type MintBankKeeper interface {
	MintCoins(ctx context.Context, moduleName string, coins sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
}

// ProvideExampleMintFn returns the function used in x/mint's endblocker to mint new tokens.
// Note that this function can not have the mint keeper as a parameter because it would create a cyclic dependency.
func ProvideExampleMintFn(bankKeeper MintBankKeeper) minttypes.MintFn {
	return func(ctx context.Context, env appmodule.Environment, minter *minttypes.Minter, epochID string, epochNumber int64) error {
		// in this example we ignore epochNumber as we don't care what epoch we are in, we just assume we are being called every minute.
		if epochID != "minute" {
			return nil
		}

		resp, err := env.QueryRouterService.Invoke(ctx, &stakingtypes.QueryParamsRequest{})
		if err != nil {
			return err
		}
		stakingParams, ok := resp.(*stakingtypes.QueryParamsResponse)
		if !ok {
			return fmt.Errorf("unexpected response type: %T", resp)
		}

		resp, err = env.QueryRouterService.Invoke(ctx, &banktypes.QuerySupplyOfRequest{Denom: stakingParams.Params.BondDenom})
		if err != nil {
			return err
		}
		bankSupply, ok := resp.(*banktypes.QuerySupplyOfResponse)
		if !ok {
			return fmt.Errorf("unexpected response type: %T", resp)
		}

		stakingTokenSupply := bankSupply.Amount

		resp, err = env.QueryRouterService.Invoke(ctx, &minttypes.QueryParamsRequest{})
		if err != nil {
			return err
		}
		mintParams, ok := resp.(*minttypes.QueryParamsResponse)
		if !ok {
			return fmt.Errorf("unexpected response type: %T", resp)
		}

		resp, err = env.QueryRouterService.Invoke(ctx, &stakingtypes.QueryPoolRequest{})
		if err != nil {
			return err
		}

		stakingPool, ok := resp.(*stakingtypes.QueryPoolResponse)
		if !ok {
			return fmt.Errorf("unexpected response type: %T", resp)
		}

		// bondedRatio
		bondedRatio := math.LegacyNewDecFromInt(stakingPool.Pool.BondedTokens).QuoInt(stakingTokenSupply.Amount)
		minter.Inflation = minter.NextInflationRate(mintParams.Params, bondedRatio)
		minter.AnnualProvisions = minter.NextAnnualProvisions(mintParams.Params, stakingTokenSupply.Amount)

		// to get a more accurate amount of tokens minted, we get, and later store, last minting time.
		// if this is the first time minting, we initialize the minter.Data with the current time - 60s
		// to mint tokens at the beginning. Note: this is a custom behavior to avoid breaking tests.
		if minter.Data == nil {
			minter.Data = make([]byte, 8)
			binary.BigEndian.PutUint64(minter.Data, (uint64)(env.HeaderService.HeaderInfo(ctx).Time.UnixMilli()-60000))
		}

		lastMint := binary.BigEndian.Uint64(minter.Data)
		binary.BigEndian.PutUint64(minter.Data, (uint64)(env.HeaderService.HeaderInfo(ctx).Time.UnixMilli()))

		// calculate the amount of tokens to mint, based on the time since the last mint.
		msSinceLastMint := env.HeaderService.HeaderInfo(ctx).Time.UnixMilli() - (int64)(lastMint)
		provisionAmt := minter.AnnualProvisions.QuoInt64(31536000000).MulInt64(msSinceLastMint) // 31536000000 = milliseconds in a year
		mintedCoin := sdk.NewCoin(mintParams.Params.MintDenom, provisionAmt.TruncateInt())
		maxSupply := mintParams.Params.MaxSupply
		totalSupply := stakingTokenSupply.Amount

		if !maxSupply.IsZero() {
			// supply is not infinite, check the amount to mint
			remainingSupply := maxSupply.Sub(totalSupply)

			if remainingSupply.LTE(math.ZeroInt()) {
				// max supply reached, no new tokens will be minted
				// also handles the case where totalSupply > maxSupply
				return nil
			}

			// if the amount to mint is greater than the remaining supply, mint the remaining supply
			if mintedCoin.Amount.GT(remainingSupply) {
				mintedCoin.Amount = remainingSupply
			}
		}

		if mintedCoin.Amount.IsZero() {
			// skip as no coins need to be minted
			return nil
		}

		mintedCoins := sdk.NewCoins(mintedCoin)
		if err := bankKeeper.MintCoins(ctx, minttypes.ModuleName, mintedCoins); err != nil {
			return err
		}

		// Example of custom send while minting
		// Send some tokens to a "team account"
		// if err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, ... ); err != nil {
		// 	return err
		// }

		if err = bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, mintedCoins); err != nil {
			return err
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
