package types

import (
	"bytes"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// msg types
const (
	TypeMsgTransfer = "transfer"
)

// NewMsgTransfer creates a new MsgTransfer instance
func NewMsgTransfer(
	sourcePort, sourceChannel string,
	amount sdk.Coins, trace DenomTrace,
	sender sdk.AccAddress, receiver string,
	timeoutHeight, timeoutTimestamp uint64,
) *MsgTransfer {
	return &MsgTransfer{
		SourcePort:       sourcePort,
		SourceChannel:    sourceChannel,
		Amount:           amount,
		Trace:            trace,
		Sender:           sender,
		Receiver:         receiver,
		TimeoutHeight:    timeoutHeight,
		TimeoutTimestamp: timeoutTimestamp,
	}
}

// Route implements sdk.Msg
func (MsgTransfer) Route() string {
	return RouterKey
}

// Type implements sdk.Msg
func (MsgTransfer) Type() string {
	return TypeMsgTransfer
}

// ValidateBasic performs a basic check of the MsgTransfer fields.
// NOTE: timeout height and timestamp values can be 0 to disable the timeout.
func (msg MsgTransfer) ValidateBasic() error {
	if err := host.PortIdentifierValidator(msg.SourcePort); err != nil {
		return sdkerrors.Wrap(err, "invalid source port ID")
	}
	if err := host.ChannelIdentifierValidator(msg.SourceChannel); err != nil {
		return sdkerrors.Wrap(err, "invalid source channel ID")
	}
	if !msg.Amount.IsAllPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFunds, msg.Amount.String())
	}
	if !msg.Amount.IsValid() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, msg.Amount.String())
	}
	if err := msg.Trace.Validate(); err != nil {
		return err
	}
	// Only validate the ibc denomination when trace info is not provided
	if msg.Trace.Trace != "" {
		denomTraceHash, err := ValidateIBCDenom(msg.Amount[0].Denom)
		if err != nil {
			return err
		}
		traceHash := msg.Trace.Hash()
		if !bytes.Equal(traceHash.Bytes(), denomTraceHash.Bytes()) {
			return fmt.Errorf("token denomination trace hash mismatch, expected %s got %s", traceHash, denomTraceHash)
		}
	} else if msg.Trace.BaseDenom != msg.Amount[0].Denom {
		// otherwise, validate that denominations are equal
		return sdkerrors.Wrapf(
			ErrInvalidDenomForTransfer,
			"token denom must match the trace base denom (%s ≠ %s)",
			msg.Amount, msg.Trace.BaseDenom,
		)
	}
	if msg.Sender.Empty() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing sender address")
	}
	if msg.Receiver == "" {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "missing recipient address")
	}
	return nil
}

// GetSignBytes implements sdk.Msg
func (msg MsgTransfer) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners implements sdk.Msg
func (msg MsgTransfer) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
