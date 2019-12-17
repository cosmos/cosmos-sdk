package types

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// DONTCOVER

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Params   Params              `json:"params" yaml:"params"`
	Evidence []exported.Evidence `json:"evidence" yaml:"evidence"`
}

func NewGenesisState(p Params, e []exported.Evidence) GenesisState {
	return GenesisState{
		Params:   p,
		Evidence: e,
	}
}

// DefaultGenesisState returns the evidence module's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params:   DefaultParams(),
		Evidence: []exported.Evidence{},
	}
}

// Validate performs basic gensis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, e := range gs.Evidence {
		if err := e.ValidateBasic(); err != nil {
			return err
		}
	}

	maxEvidence := gs.Params.MaxEvidenceAge
	if maxEvidence < 1*time.Minute {
		return fmt.Errorf("max evidence age must be at least 1 minute, is %s", maxEvidence.String())
	}

	return nil
}
