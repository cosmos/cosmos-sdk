package types

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameter namespace
const (
	DefaultParamspace                         = ModuleName
	DefaultMaxEvidenceAge       time.Duration = 60 * 2 * time.Second
	DefaultSignedBlocksWindow   int64         = 100
	DefaultDowntimeJailDuration time.Duration = 60 * 10 * time.Second
)

// The Double Sign Jail period ends at Max Time supported by Amino (Dec 31, 9999 - 23:59:59 GMT)
var (
	DoubleSignJailEndTime          = time.Unix(253402300799, 0)
	DefaultMinSignedPerWindow      = sdk.NewDecWithPrec(5, 1)
	DefaultSlashFractionDoubleSign = sdk.NewDec(1).Quo(sdk.NewDec(20))
	DefaultSlashFractionDowntime   = sdk.NewDec(1).Quo(sdk.NewDec(100))
)

// Parameter store keys
var (
	KeyMaxEvidenceAge          = []byte("MaxEvidenceAge")
	KeySignedBlocksWindow      = []byte("SignedBlocksWindow")
	KeyMinSignedPerWindow      = []byte("MinSignedPerWindow")
	KeyDowntimeJailDuration    = []byte("DowntimeJailDuration")
	KeySlashFractionDoubleSign = []byte("SlashFractionDoubleSign")
	KeySlashFractionDowntime   = []byte("SlashFractionDowntime")
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
	SlashFractionDowntime   sdk.Dec       `json:"slash_fraction_downtime"`
}

// NewParams creates a new Params object
func NewParams(maxEvidenceAge time.Duration, signedBlocksWindow int64,
	minSignedPerWindow sdk.Dec, downtimeJailDuration time.Duration,
	slashFractionDoubleSign sdk.Dec, slashFractionDowntime sdk.Dec) Params {

	return Params{
		MaxEvidenceAge:          maxEvidenceAge,
		SignedBlocksWindow:      signedBlocksWindow,
		MinSignedPerWindow:      minSignedPerWindow,
		DowntimeJailDuration:    downtimeJailDuration,
		SlashFractionDoubleSign: slashFractionDoubleSign,
		SlashFractionDowntime:   slashFractionDowntime,
	}
}

func (p Params) String() string {
	return fmt.Sprintf(`Slashing Params:
  MaxEvidenceAge:          %s
  SignedBlocksWindow:      %d
  MinSignedPerWindow:      %s
  DowntimeJailDuration:    %s
  SlashFractionDoubleSign: %s
  SlashFractionDowntime:   %s`, p.MaxEvidenceAge,
		p.SignedBlocksWindow, p.MinSignedPerWindow,
		p.DowntimeJailDuration, p.SlashFractionDoubleSign,
		p.SlashFractionDowntime)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMaxEvidenceAge, &p.MaxEvidenceAge},
		{KeySignedBlocksWindow, &p.SignedBlocksWindow},
		{KeyMinSignedPerWindow, &p.MinSignedPerWindow},
		{KeyDowntimeJailDuration, &p.DowntimeJailDuration},
		{KeySlashFractionDoubleSign, &p.SlashFractionDoubleSign},
		{KeySlashFractionDowntime, &p.SlashFractionDowntime},
	}
}

// Default parameters for this module
func DefaultParams() Params {
	return Params{
		MaxEvidenceAge:          DefaultMaxEvidenceAge,
		SignedBlocksWindow:      DefaultSignedBlocksWindow,
		MinSignedPerWindow:      DefaultMinSignedPerWindow,
		DowntimeJailDuration:    DefaultDowntimeJailDuration,
		SlashFractionDoubleSign: DefaultSlashFractionDoubleSign,
		SlashFractionDowntime:   DefaultSlashFractionDowntime,
	}
}
