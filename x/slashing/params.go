package slashing

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace = "slashing"
)

// The Double Sign Jail period ends at Max Time supported by Amino (Dec 31, 9999 - 23:59:59 GMT)
var (
	DoubleSignJailEndTime = time.Unix(253402300799, 0)
)

// Parameter store key
var (
	KeyMaxEvidenceAge          = []byte("MaxEvidenceAge")
	KeySignedBlocksWindow      = []byte("SignedBlocksWindow")
	KeyMinSignedPerWindow      = []byte("MinSignedPerWindow")
	KeyDowntimeJailDuration    = []byte("DowntimeJailDuration")
	KeySlashFractionDoubleSign = []byte("SlashFractionDoubleSign")
)

// ParamKeyTable for slashing module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// Params - used for initializing default parameter for slashing at genesis
type Params struct {
	MaxEvidenceAge          time.Duration `json:"max_evidence_age"`
	SignedBlocksWindow      int64         `json:"signed_blocks_window"`
	MinSignedPerWindow      sdk.Dec       `json:"min_signed_per_window"`
	DowntimeJailDuration    time.Duration `json:"downtime_jail_duration"`
	SlashFractionDoubleSign sdk.Dec       `json:"slash_fraction_double_sign"`
}

func (p Params) String() string {
	return fmt.Sprintf(`Slashing Params:
  MaxEvidenceAge:          %s
  SignedBlocksWindow:      %d
  MinSignedPerWindow:      %s
  DowntimeJailDuration:    %s
  SlashFractionDoubleSign: %d`, p.MaxEvidenceAge,
		p.SignedBlocksWindow, p.MinSignedPerWindow,
		p.DowntimeJailDuration, p.SlashFractionDoubleSign)
}

// Implements params.ParamStruct
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMaxEvidenceAge, &p.MaxEvidenceAge},
		{KeySignedBlocksWindow, &p.SignedBlocksWindow},
		{KeyMinSignedPerWindow, &p.MinSignedPerWindow},
		{KeyDowntimeJailDuration, &p.DowntimeJailDuration},
		{KeySlashFractionDoubleSign, &p.SlashFractionDoubleSign},
	}
}

// Default parameters used by Cosmos Hub
func DefaultParams() Params {
	return Params{
		// defaultMaxEvidenceAge = 60 * 60 * 24 * 7 * 3
		// Set to 2 minutes for testnets.
		MaxEvidenceAge: 60 * 2 * time.Second,

		// Set to 100 blocks for testnets
		SignedBlocksWindow: 100,

		// Set to 10 minutes for testnets
		DowntimeJailDuration: 60 * 10 * time.Second,

		// CONTRACT must be less than 1
		// TODO enforce this contract https://github.com/cosmos/cosmos-sdk/issues/3474
		// Set to 10%, viable for both testnets & mainnets
		MinSignedPerWindow: sdk.NewDecWithPrec(1, 1),

		// Set to 5% for testnets
		SlashFractionDoubleSign: sdk.NewDec(1).Quo(sdk.NewDec(20)),
	}
}

// MaxEvidenceAge - Max age for evidence - 21 days (3 weeks)
// MaxEvidenceAge = 60 * 60 * 24 * 7 * 3
func (k Keeper) MaxEvidenceAge(ctx sdk.Context) (res time.Duration) {
	k.paramspace.Get(ctx, KeyMaxEvidenceAge, &res)
	return
}

// SignedBlocksWindow - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Context) (res int64) {
	k.paramspace.Get(ctx, KeySignedBlocksWindow, &res)
	return
}

// Downtime slashing threshold - default 50% of the SignedBlocksWindow
func (k Keeper) MinSignedPerWindow(ctx sdk.Context) int64 {
	var minSignedPerWindow sdk.Dec
	k.paramspace.Get(ctx, KeyMinSignedPerWindow, &minSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)

	// NOTE: RoundInt64 will never panic as minSignedPerWindow is
	//       less than 1.
	return minSignedPerWindow.MulInt64(signedBlocksWindow).RoundInt64()
}

// Downtime unbond duration
func (k Keeper) DowntimeJailDuration(ctx sdk.Context) (res time.Duration) {
	k.paramspace.Get(ctx, KeyDowntimeJailDuration, &res)
	return
}

// SlashFractionDoubleSign - currently default 5%
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Context) (res sdk.Dec) {
	k.paramspace.Get(ctx, KeySlashFractionDoubleSign, &res)
	return
}

// GetParams returns the total set of slashing parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params Params) {
	k.paramspace.GetParamSet(ctx, &params)
	return params
}
