// DONTCOVER
package types

// GenesisState defines the evidence module's genesis state.
type GenesisState struct {
	Evidence []Evidence `json:"evidence" yaml:"evidence"`
}

func NewGenesisState(e []Evidence) GenesisState {
	return GenesisState{Evidence: e}
}

// DefaultGenesisState returns the evidence module's default genesis state.
func DefaultGenesisState() GenesisState {
	return GenesisState{Evidence: []Evidence{}}
}

// Validate performs basic gensis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	for _, e := range gs.Evidence {
		if err := e.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}
