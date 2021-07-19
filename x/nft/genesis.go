package nft

// ValidateGenesis check the given genesis state has no integrity issues
func ValidateGenesis(data GenesisState) error {
	for _, class := range data.Classes {
		if err := ValidateClassID(class.Id); err != nil {
			panic(err)
		}
	}
	for _, entry := range data.Entries {
		for _, nft := range entry.NFTs {
			if err := ValidateNFTID(nft.Id); err != nil {
				panic(err)
			}
		}
	}
	return nil
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}
