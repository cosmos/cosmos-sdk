package slashing

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint
const (
	MaxEvidenceAgeKey           = "slashing/MaxEvidenceAge"
	SignedBlocksWindowKey       = "slashing/SignedBlocksWindow"
	MinSignedPerWindowKey       = "slashing/MinSignedPerWindow"
	DoubleSignUnbondDurationKey = "slashing/DoubleSignUnbondDuration"
	DowntimeUnbondDurationKey   = "slashing/DowntimeUnbondDuration"
	SlashFractionDoubleSignKey  = "slashing/SlashFractionDoubleSign"
	SlashFractionDowntimeKey    = "slashing/SlashFractionDowntime"
)

// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
func (k Keeper) MaxEvidenceAge(ctx sdk.Context) time.Duration {
	return time.Duration(k.params.GetInt64WithDefault(ctx, MaxEvidenceAgeKey, defaultMaxEvidenceAge)) * time.Second
}

// SignedBlocksWindow - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Context) int64 {
	return k.params.GetInt64WithDefault(ctx, SignedBlocksWindowKey, defaultSignedBlocksWindow)
}

// Downtime slashing thershold - default 50%
func (k Keeper) MinSignedPerWindow(ctx sdk.Context) int64 {
	minSignedPerWindow := k.params.GetRatWithDefault(ctx, MinSignedPerWindowKey, defaultMinSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)
	return sdk.NewRat(signedBlocksWindow).Mul(minSignedPerWindow).RoundInt64()
}

// Double-sign unbond duration
func (k Keeper) DoubleSignUnbondDuration(ctx sdk.Context) time.Duration {
	return time.Duration(k.params.GetInt64WithDefault(ctx, DoubleSignUnbondDurationKey, defaultDoubleSignUnbondDuration)) * time.Second
}

// Downtime unbond duration
func (k Keeper) DowntimeUnbondDuration(ctx sdk.Context) time.Duration {
	return time.Duration(k.params.GetInt64WithDefault(ctx, DowntimeUnbondDurationKey, defaultDowntimeUnbondDuration)) * time.Second
}

// SlashFractionDoubleSign - currently default 5%
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Context) sdk.Rat {
	return k.params.GetRatWithDefault(ctx, SlashFractionDoubleSignKey, defaultSlashFractionDoubleSign)
}

// SlashFractionDowntime - currently default 1%
func (k Keeper) SlashFractionDowntime(ctx sdk.Context) sdk.Rat {
	return k.params.GetRatWithDefault(ctx, SlashFractionDowntimeKey, defaultSlashFractionDowntime)
}

// declared as var because of keeper_test.go
// TODO: make it const or parameter of NewKeeper

var (
	// defaultMaxEvidenceAge = 60 * 60 * 24 * 7 * 3
	// TODO Temporarily set to 2 minutes for testnets.
	defaultMaxEvidenceAge int64 = 60 * 2

	// TODO Temporarily set to five minutes for testnets
	defaultDoubleSignUnbondDuration int64 = 60 * 5

	// TODO Temporarily set to 10000 blocks for testnets
	defaultSignedBlocksWindow int64 = 10000

	// TODO Temporarily set to 10 minutes for testnets
	defaultDowntimeUnbondDuration int64 = 60 * 10

	defaultMinSignedPerWindow = sdk.NewRat(1, 2)

	defaultSlashFractionDoubleSign = sdk.NewRat(1).Quo(sdk.NewRat(20))

	defaultSlashFractionDowntime = sdk.NewRat(1).Quo(sdk.NewRat(100))
)
