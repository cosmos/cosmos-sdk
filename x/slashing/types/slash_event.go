package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// slashing message types
const (
	TypeSlashEvent = "slash_event"
)

// verify interface at compile time
var _ sdk.Msg = &SlashEvent{}

// NewSlashEvent creates a new SlashEvent instance
func NewSlashEvent(address sdk.ValAddress, votingPercent sdk.Dec, slashPercent sdk.Dec, height int64, power int64) *SlashEvent {
	return &SlashEvent{
		Address:                address.String(),
		ValidatorVotingPercent: votingPercent,
		SlashPercent:           slashPercent,
		DistributionHeight:     height,
		ValidatorPower:         power,
	}
}

func (msg SlashEvent) Route() string { return RouterKey }
func (msg SlashEvent) Type() string  { return TypeSlashEvent }
func (msg SlashEvent) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

// GetSignBytes gets the bytes for the message signer to sign on
func (msg SlashEvent) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic validity check for the AnteHandler
func (msg SlashEvent) ValidateBasic() error {
	return nil
}
