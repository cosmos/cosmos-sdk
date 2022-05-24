package v039

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "slashing"
)

// Default parameter namespace
const (
	DefaultParamspace           = ModuleName
	DefaultSignedBlocksWindow   = int64(100)
	DefaultDowntimeJailDuration = 60 * 10 * time.Second
)

var (
	DefaultMinSignedPerWindow      = sdk.NewDecWithPrec(5, 1)
	DefaultSlashFractionDoubleSign = sdk.NewDec(1).Quo(sdk.NewDec(20))
	DefaultSlashFractionDowntime   = sdk.NewDec(1).Quo(sdk.NewDec(100))
)

// Params - used for initializing default parameter for slashing at genesis
type Params struct {
	SignedBlocksWindow      int64         `json:"signed_blocks_window" yaml:"signed_blocks_window"`
	MinSignedPerWindow      sdk.Dec       `json:"min_signed_per_window" yaml:"min_signed_per_window"`
	DowntimeJailDuration    time.Duration `json:"downtime_jail_duration" yaml:"downtime_jail_duration"`
	SlashFractionDoubleSign sdk.Dec       `json:"slash_fraction_double_sign" yaml:"slash_fraction_double_sign"`
	SlashFractionDowntime   sdk.Dec       `json:"slash_fraction_downtime" yaml:"slash_fraction_downtime"`
}

// NewParams creates a new Params object
func NewParams(
	signedBlocksWindow int64, minSignedPerWindow sdk.Dec, downtimeJailDuration time.Duration,
	slashFractionDoubleSign, slashFractionDowntime sdk.Dec,
) Params {

	return Params{
		SignedBlocksWindow:      signedBlocksWindow,
		MinSignedPerWindow:      minSignedPerWindow,
		DowntimeJailDuration:    downtimeJailDuration,
		SlashFractionDoubleSign: slashFractionDoubleSign,
		SlashFractionDowntime:   slashFractionDowntime,
	}
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return NewParams(
		DefaultSignedBlocksWindow, DefaultMinSignedPerWindow, DefaultDowntimeJailDuration,
		DefaultSlashFractionDoubleSign, DefaultSlashFractionDowntime,
	)
}

// ValidatorSigningInfo defines the signing info for a validator
type ValidatorSigningInfo struct {
	Address             sdk.ConsAddress `json:"address" yaml:"address"`                             // validator consensus address
	StartHeight         int64           `json:"start_height" yaml:"start_height"`                   // height at which validator was first a candidate OR was unjailed
	IndexOffset         int64           `json:"index_offset" yaml:"index_offset"`                   // index offset into signed block bit array
	JailedUntil         time.Time       `json:"jailed_until" yaml:"jailed_until"`                   // timestamp validator cannot be unjailed until
	Tombstoned          bool            `json:"tombstoned" yaml:"tombstoned"`                       // whether or not a validator has been tombstoned (killed out of validator set)
	MissedBlocksCounter int64           `json:"missed_blocks_counter" yaml:"missed_blocks_counter"` // missed blocks counter (to avoid scanning the array every time)
}

// MissedBlock
type MissedBlock struct {
	Index  int64 `json:"index" yaml:"index"`
	Missed bool  `json:"missed" yaml:"missed"`
}

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params       Params                          `json:"params" yaml:"params"`
	SigningInfos map[string]ValidatorSigningInfo `json:"signing_infos" yaml:"signing_infos"`
	MissedBlocks map[string][]MissedBlock        `json:"missed_blocks" yaml:"missed_blocks"`
}
