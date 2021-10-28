package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateGenesis check the given genesis state has no integrity issues
func ValidateGenesis(data GenesisState) error {
	for _, class := range data.Classes {
		if err := ValidateClassID(class.Id); err != nil {
			return err
		}
	}
	for _, entry := range data.Entries {
		for _, nft := range entry.Nfts {
			if err := ValidateNFTID(nft.Id); err != nil {
				return err
			}
			if _, err := sdk.AccAddressFromBech32(entry.Owner); err != nil {
				return err
			}
		}
	}
	return nil
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}
