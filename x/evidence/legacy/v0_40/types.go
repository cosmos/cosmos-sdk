package v040

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v038evidence "github.com/cosmos/cosmos-sdk/x/evidence/legacy/v0_38"
)

// DONTCOVER
// nolint

const (
	ModuleName = "evidence"
)

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Evidence []*codectypes.Any `protobuf:"bytes,1,rep,name=evidence,proto3" json:"evidence,omitempty"`
}

// NewGenesisState creates a new genesis state for the evidence module.
func NewGenesisState(e []v038evidence.Evidence) GenesisState {
	evidence := make([]*codectypes.Any, len(e))
	for i, evi := range e {
		msg, ok := evi.(proto.Message)
		if !ok {
			panic(fmt.Errorf("cannot proto marshal %T", evi))
		}
		any, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
		evidence[i] = any
	}
	return GenesisState{
		Evidence: evidence,
	}
}
