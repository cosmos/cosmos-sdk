package nft

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ValidateGenesis check the given genesis state has no integrity issues
func ValidateGenesis(data GenesisState) error {
	for _, class := range data.Classes {
		if len(class.Id) == 0 {
			return errors.Wrapf(ErrInvalidID, "Empty class id (%s)", class.Id)
		}
	}
	for _, entry := range data.Entries {
		for _, nft := range entry.Nfts {
			if len(nft.Id) == 0 {
				return errors.Wrapf(ErrInvalidID, "Empty nft id (%s)", nft.Id)
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
