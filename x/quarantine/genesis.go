package quarantine

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Validate performs basic validation of genesis data returning an error for any failed validation criteria.
func (gs GenesisState) Validate() error {
	for i, addr := range gs.QuarantinedAddresses {
		if _, err := sdk.AccAddressFromBech32(addr); err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("invalid quarantined address[%d]: %v", i, err)
		}
	}
	for i, resp := range gs.AutoResponses {
		if err := resp.Validate(); err != nil {
			return errors.Wrapf(err, "invalid quarantine auto response entry[%d]", i)
		}
	}
	for i, funds := range gs.QuarantinedFunds {
		if err := funds.Validate(); err != nil {
			return errors.Wrapf(err, "invalid quarantined funds[%d]", i)
		}
	}
	return nil
}

// NewGenesisState creates a new genesis state for the quarantine module.
func NewGenesisState(quarantinedAddresses []string, autoResponses []*AutoResponseEntry, funds []*QuarantinedFunds) *GenesisState {
	return &GenesisState{
		QuarantinedAddresses: quarantinedAddresses,
		AutoResponses:        autoResponses,
		QuarantinedFunds:     funds,
	}
}

// DefaultGenesisState returns a default quarantine module genesis state.
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(nil, nil, nil)
}
