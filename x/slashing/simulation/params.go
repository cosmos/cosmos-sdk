package simulation

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Simulation parameter constants
const (
	SignedBlocksWindow      = "signed_blocks_window"
	MinSignedPerWindow      = "min_signed_per_window"
	DowntimeJailDuration    = "downtime_jail_duration"
	SlashFractionDoubleSign = "slash_fraction_double_sign"
	SlashFractionDowntime   = "slash_fraction_downtime"
)

// GenParams generates random gov parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[SignedBlocksWindow] = func(r *rand.Rand) interface{} {
		return int64(RandIntBetween(r, 10, 1000))
	}

	paramSims[MinSignedPerWindow] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(int64(r.Intn(10)), 1)
	}

	paramSims[DowntimeJailDuration] = func(r *rand.Rand) interface{} {
		return time.Duration(RandIntBetween(r, 60, 60*60*24)) * time.Second
	}

	paramSims[SlashFractionDoubleSign] = func(r *rand.Rand) interface{} {
		return sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(50) + 1)))
	}

	paramSims[SlashFractionDowntime] = func(r *rand.Rand) interface{} {
		return sdk.NewDec(1).Quo(sdk.NewDec(int64(r.Intn(200) + 1)))
	}
}
