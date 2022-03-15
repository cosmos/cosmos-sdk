package types

// Validate performs basic validation of supply genesis data returning an
// error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	// TODO
	return nil
}

// DefaultGenesisState returns a default tieredfee module genesis state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		ParentGasUsed: 0,
		GasPrices:     nil,
	}
}
