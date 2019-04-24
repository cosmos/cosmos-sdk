package types

// GenesisState is the supply state that must be provided at genesis.
type GenesisState struct {
	Supplier Supplier `json:"supplier"`
}

// NewGenesisState creates a new genesis state.
func NewGenesisState(supplier Supplier) GenesisState {
	return GenesisState{Supplier: supplier}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() GenesisState {
	return NewGenesisState(DefaultSupplier())
}
