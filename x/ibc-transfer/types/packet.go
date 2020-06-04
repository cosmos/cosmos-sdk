package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// DefaultRelativePacketTimeoutHeight is the default packet timeout height (in blocks) relative
	// to the current block height. The timeout is disabled when set to 0.
	DefaultRelativePacketTimeoutHeight = 1000

	// DefaultRelativePacketTimeoutTimestamp is the default packet timeout timestamp (in nanoseconds)
	// relative to the current block timestamp. The timeout is disabled when set to 0.
	DefaultRelativePacketTimeoutTimestamp = 0

	// DefaultAbsolutePacketTimeoutHeight is the default packet timeout in blocks.
	// The timeout is disabled when set to 0.
	DefaultAbsolutePacketTimeoutHeight = 0

	// DefaultAbsolutePacketTimeoutTimestamp is the default packet timeout timestamp (in nanoseconds)
	// relative to the current block timestamp. The timeout is disabled when set to 0.
	DefaultAbsolutePacketTimeoutTimestamp = 0
)

// NewFungibleTokenPacketData contructs a new FungibleTokenPacketData instance
func NewFungibleTokenPacketData(
	amount sdk.Coins, sender, receiver string) FungibleTokenPacketData {
	return FungibleTokenPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
	}
}

// ValidateBasic is used for validating the token transfer
func (ftpd FungibleTokenPacketData) ValidateBasic() error {
	if !ftpd.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, ftpd.Amount.String())
	}
	if !ftpd.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, ftpd.Amount.String())
	}
	if strings.TrimSpace(ftpd.Sender) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "sender address cannot be blank")
	}
	if strings.TrimSpace(ftpd.Receiver) == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "receiver address cannot be blank")
	}
	return nil
}

// GetBytes is a helper for serialising
func (ftpd FungibleTokenPacketData) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ftpd))
}

// GetBytes is a helper for serialising
func (ack FungibleTokenPacketAcknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(ack))
}
