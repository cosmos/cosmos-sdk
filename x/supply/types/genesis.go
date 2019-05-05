package types

 // GenesisState is the supply state that must be provided at genesis.
type GenesisState struct {
	Supply Supply `json:"supply"`
	SendEnabled bool `json:"send_enabled"`
}

 // NewGenesisState creates a new genesis state.
func NewGenesisState(supply Supply, sendEnabled bool) GenesisState {
	return GenesisState{supply, sendEnabled}
}

 // DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultSupply(), false)
}