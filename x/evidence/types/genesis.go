package types

import (
	"fmt"

	"github.com/cosmos/gogoproto/proto"

	"cosmossdk.io/x/evidence/exported"

	codectypes "github.com/cosmos/gogoproto/types/any"
)

var _ codectypes.UnpackInterfacesMessage = GenesisState{}

// NewGenesisState creates a new genesis state for the evidence module.
func NewGenesisState(e []exported.Evidence) *GenesisState {
	evidence := make([]*codectypes.Any, len(e))
	for i, evi := range e {
		msg, ok := evi.(proto.Message)
		if !ok {
			panic(fmt.Errorf("cannot proto marshal %T", evi))
		}
		value, err := codectypes.NewAnyWithCacheWithValue(msg)
		if err != nil {
			panic(err)
		}
		evidence[i] = value
	}
	return &GenesisState{
		Evidence: evidence,
	}
}

// DefaultGenesisState returns the evidence module's default genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Evidence: []*codectypes.Any{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
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
func (gs GenesisState) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, value := range gs.Evidence {
		var evi exported.Evidence
		err := unpacker.UnpackAny(value, &evi)
		if err != nil {
			return err
		}
	}
	return nil
}
