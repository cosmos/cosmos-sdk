package types

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

var _ types.UnpackInterfacesMessage = GenesisState{}

// NewGenesisState creates a new genesis state for the evidence module.
func NewGenesisState(e []exported.Evidence) *GenesisState {
	evidence := make([]*types.Any, len(e))
	for i, evi := range e {
		msg, ok := evi.(proto.Message)
		if !ok {
			panic(fmt.Errorf("cannot proto marshal %T", evi))
		}
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			panic(err)
		}
		evidence[i] = any
	}
	return &GenesisState{
		Evidence: evidence,
	}
}

// DefaultGenesisState returns the evidence module's default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Evidence: []*types.Any{},
	}
}

// Validate performs basic gensis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, e := range gs.Evidence {
		evi, ok := e.GetCachedValue().(exported.Evidence)
		if !ok {
			return fmt.Errorf("expected evidence")
		}
		if err := evi.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (gs GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, any := range gs.Evidence {
		var evi exported.Evidence
		err := unpacker.UnpackAny(any, &evi)
		if err != nil {
			return err
		}
	}
	return nil
}
