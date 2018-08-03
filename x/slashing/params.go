package slashing

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamSpace = "Slashing"
)

// nolint - Key generators for parameter access
func MaxEvidenceAgeKey() params.Key           { return params.NewKey("MaxEvidenceAge") }
func SignedBlocksWindowKey() params.Key       { return params.NewKey("SignedBlocksWindow") }
func MinSignedPerWindowKey() params.Key       { return params.NewKey("MinSignedPerWindow") }
func DoubleSignUnbondDurationKey() params.Key { return params.NewKey("DoubleSignUnbondDuration") }
func DowntimeUnbondDurationKey() params.Key   { return params.NewKey("DowntimeUnbondDuration") }
func SlashFractionDoubleSignKey() params.Key  { return params.NewKey("SlashFractionDoubleSign") }
func SlashFractionDowntimeKey() params.Key    { return params.NewKey("SlashFractionDowntime") }

// Cached parameter keys
var (
	maxEvidenceAgeKey           = MaxEvidenceAgeKey()
	signedBlocksWindowKey       = SignedBlocksWindowKey()
	minSignedPerWindowKey       = MinSignedPerWindowKey()
	doubleSignUnbondDurationKey = DoubleSignUnbondDurationKey()
	downtimeUnbondDurationKey   = DowntimeUnbondDurationKey()
	slashFractionDoubleSignKey  = SlashFractionDoubleSignKey()
	slashFractionDowntimeKey    = SlashFractionDowntimeKey()
)

// Params - used for initializing default parameter for slashing at genesis
type Params struct {
	MaxEvidenceAge           int64   `json:"max-evidence-age"`
	SignedBlocksWindow       int64   `json:"signed-blocks-window"`
	MinSignedPerWindow       sdk.Dec `json:"min-signed-per-window"`
	DoubleSignUnbondDuration int64   `json:"doublesign-unbond-duration"`
	DowntimeUnbondDuration   int64   `json:"downtime-unbond-duration"`
	SlashFractionDoubleSign  sdk.Dec `json:"slash-fraction-doublesign"`
	SlashFractionDowntime    sdk.Dec `json:"slash-fraction-downtime"`
}

// Default parameters used by Cosmos Hub
func HubDefaultParams() Params {
	return Params{
		// defaultMaxEvidenceAge = 60 * 60 * 24 * 7 * 3
		// TODO Temporarily set to 2 minutes for testnets.
		MaxEvidenceAge: 60 * 2,

		// TODO Temporarily set to five minutes for testnets
		DoubleSignUnbondDuration: 60 * 5,

		// TODO Temporarily set to 100 blocks for testnets
		SignedBlocksWindow: 100,

		// TODO Temporarily set to 10 minutes for testnets
		DowntimeUnbondDuration: 60 * 10,

		MinSignedPerWindow: sdk.NewDecWithPrec(5, 1),

		SlashFractionDoubleSign: sdk.NewDec(1).Quo(sdk.NewDec(20)),

		SlashFractionDowntime: sdk.NewDec(1).Quo(sdk.NewDec(100)),
	}
}

// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
func (k Keeper) MaxEvidenceAge(ctx sdk.Context) time.Duration {
	var t int64
	k.paramstore.Get(ctx, maxEvidenceAgeKey, &t)
	return time.Duration(t) * time.Second
}

// SignedBlocksWindow - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Context) (res int64) {
	k.paramstore.Get(ctx, signedBlocksWindowKey, &res)
	return
}

// Downtime slashing thershold - default 50% of the SignedBlocksWindow
func (k Keeper) MinSignedPerWindow(ctx sdk.Context) int64 {
	var minSignedPerWindow sdk.Dec
	k.paramstore.Get(ctx, minSignedPerWindowKey, &minSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)
	return sdk.NewDec(signedBlocksWindow).Mul(minSignedPerWindow).RoundInt64()
}

// Double-sign unbond duration
func (k Keeper) DoubleSignUnbondDuration(ctx sdk.Context) time.Duration {
	var t int64
	k.paramstore.Get(ctx, doubleSignUnbondDurationKey, &t)
	return time.Duration(t) * time.Second
}

// Downtime unbond duration
func (k Keeper) DowntimeUnbondDuration(ctx sdk.Context) time.Duration {
	var t time.Duration
	k.paramstore.Get(ctx, downtimeUnbondDurationKey, &t)
	return time.Duration(t) * time.Second
}

// SlashFractionDoubleSign - currently default 5%
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, slashFractionDoubleSignKey, &res)
	return
}

// SlashFractionDowntime - currently default 1%
func (k Keeper) SlashFractionDowntime(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, slashFractionDowntimeKey, &res)
	return
}
