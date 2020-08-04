package v039

import (
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// DONTCOVER
// nolint

const (
	ModuleName = "slashing"
)

// GenesisState - all slashing state that must be provided at genesis
type GenesisState struct {
	Params       types.Params                          `json:"params" yaml:"params"`
	SigningInfos map[string]types.ValidatorSigningInfo `json:"signing_infos" yaml:"signing_infos"`
	MissedBlocks map[string][]types.MissedBlock        `json:"missed_blocks" yaml:"missed_blocks"`
}
