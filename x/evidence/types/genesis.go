package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// DONTCOVER

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Evidence []exported.Evidence `json:"evidence" yaml:"evidence"`
}

func NewGenesisState(e []exported.Evidence) GenesisState {
	return GenesisState{
		Evidence: e,
	}
}

// DefaultGenesisState returns the evidence module's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Evidence: []exported.Evidence{},
	}
}

// Validate performs basic gensis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for i, e := range gs.Evidence {
		if e == nil {
			return fmt.Errorf("evidence %d cannot be nil", i)
		}
		if err := e.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
