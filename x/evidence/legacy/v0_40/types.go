package v040

import sdk "github.com/cosmos/cosmos-sdk/codec/types"

// DONTCOVER
// nolint

const (
	ModuleName = "evidence"
)

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Evidence []*sdk.Any `protobuf:"bytes,1,rep,name=evidence,proto3" json:"evidence,omitempty"`
}
