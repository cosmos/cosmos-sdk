package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/sanction/errors"
)

func NewGenesisState(params *Params, addrs []string, tempEntries []*TemporaryEntry) *GenesisState {
	return &GenesisState{
		Params:              params,
		SanctionedAddresses: addrs,
		TemporaryEntries:    tempEntries,
	}
}

func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), nil, nil)
}

func (g GenesisState) Validate() error {
	if g.Params != nil {
		if err := g.Params.ValidateBasic(); err != nil {
			return errors.ErrInvalidParams.Wrap(err.Error())
		}
	}
	for i, addr := range g.SanctionedAddresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("sanctioned addresses[%d], %q: %v", i, addr, err)
		}
	}
	for i, entry := range g.TemporaryEntries {
		if entry.Status != TEMP_STATUS_SANCTIONED && entry.Status != TEMP_STATUS_UNSANCTIONED {
			return errors.ErrInvalidTempStatus.Wrapf("temporary entries[%d]: %s", i, entry.Status)
		}
		_, err := sdk.AccAddressFromBech32(entry.Address)
		if err != nil {
			return sdkerrors.ErrInvalidAddress.Wrapf("temporary entries[%d], %q: %v", i, entry.Address, err)
		}
	}
	return nil
}
