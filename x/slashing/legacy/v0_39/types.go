package v039

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DONTCOVER
// nolint

const (
	ModuleName = "slashing"
)

// Default parameter namespace
const (
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
	SignedBlocksWindow      int64         `protobuf:"varint,1,opt,name=signed_blocks_window,json=signedBlocksWindow,proto3" json:"signed_blocks_window,omitempty" yaml:"signed_blocks_window"`
	MinSignedPerWindow      sdk.Dec       `protobuf:"bytes,2,opt,name=min_signed_per_window,json=minSignedPerWindow,proto3,customtype=github.com/cosmos/cosmos-sdk/Dec" json:"min_signed_per_window" yaml:"min_signed_per_window"`
	DowntimeJailDuration    time.Duration `protobuf:"bytes,3,opt,name=downtime_jail_duration,json=downtimeJailDuration,proto3,stdduration" json:"downtime_jail_duration" yaml:"downtime_jail_duration"`
	SlashFractionDoubleSign sdk.Dec       `protobuf:"bytes,4,opt,name=slash_fraction_double_sign,json=slashFractionDoubleSign,proto3,customtype=github.com/cosmos/cosmos-sdk/Dec" json:"slash_fraction_double_sign" yaml:"slash_fraction_double_sign"`
	SlashFractionDowntime   sdk.Dec       `protobuf:"bytes,5,opt,name=slash_fraction_downtime,json=slashFractionDowntime,proto3,customtype=github.com/cosmos/cosmos-sdk/Dec" json:"slash_fraction_downtime" yaml:"slash_fraction_downtime"`
}

// DefaultParams defines the parameters for this module
func DefaultParams() Params {
	return Params{
		SignedBlocksWindow:      DefaultSignedBlocksWindow,
		MinSignedPerWindow:      DefaultMinSignedPerWindow,
		DowntimeJailDuration:    DefaultDowntimeJailDuration,
		SlashFractionDoubleSign: DefaultSlashFractionDoubleSign,
		SlashFractionDowntime:   DefaultSlashFractionDowntime,
	}
}

// ValidatorSigningInfo defines the signing info for a validator
type ValidatorSigningInfo struct {
	Address sdk.ConsAddress `protobuf:"bytes,1,opt,name=address,proto3,casttype=github.com/cosmos/cosmos-sdk/ConsAddress" json:"address,omitempty"`
	// height at which validator was first a candidate OR was unjailed
	StartHeight int64 `protobuf:"varint,2,opt,name=start_height,json=startHeight,proto3" json:"start_height,omitempty" yaml:"start_height"`
	// index offset into signed block bit array
	IndexOffset int64 `protobuf:"varint,3,opt,name=index_offset,json=indexOffset,proto3" json:"index_offset,omitempty" yaml:"index_offset"`
	// timestamp validator cannot be unjailed until
	JailedUntil time.Time `protobuf:"bytes,4,opt,name=jailed_until,json=jailedUntil,proto3,stdtime" json:"jailed_until" yaml:"jailed_until"`
	// whether or not a validator has been tombstoned (killed out of validator set)
	Tombstoned bool `protobuf:"varint,5,opt,name=tombstoned,proto3" json:"tombstoned,omitempty"`
	// missed blocks counter (to avoid scanning the array every time)
	MissedBlocksCounter int64 `protobuf:"varint,6,opt,name=missed_blocks_counter,json=missedBlocksCounter,proto3" json:"missed_blocks_counter,omitempty" yaml:"missed_blocks_counter"`
}

// MissedBlock contains height and missed status as boolean
type MissedBlock struct {
	Index  int64 `protobuf:"varint,1,opt,name=index,proto3" json:"index,omitempty"`
	Missed bool  `protobuf:"varint,2,opt,name=missed,proto3" json:"missed,omitempty"`
}

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params       Params                          `json:"params" yaml:"params"`
	SigningInfos map[string]ValidatorSigningInfo `json:"signing_infos" yaml:"signing_infos"`
	MissedBlocks map[string][]MissedBlock        `json:"missed_blocks" yaml:"missed_blocks"`
}
