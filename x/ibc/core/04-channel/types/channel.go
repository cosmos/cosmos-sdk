package types

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var (
	_ exported.ChannelI             = (*Channel)(nil)
	_ exported.CounterpartyChannelI = (*Counterparty)(nil)
)

// NewChannel creates a new Channel instance
func NewChannel(
	state State, ordering Order, counterparty Counterparty,
	hops []string, version string,
) Channel {
	return Channel{
		State:          state,
		Ordering:       ordering,
		Counterparty:   counterparty,
		ConnectionHops: hops,
		Version:        version,
	}
}

// GetState implements Channel interface.
func (ch Channel) GetState() int32 {
	return int32(ch.State)
}

// GetOrdering implements Channel interface.
func (ch Channel) GetOrdering() int32 {
	return int32(ch.Ordering)
}

// GetCounterparty implements Channel interface.
func (ch Channel) GetCounterparty() exported.CounterpartyChannelI {
	return ch.Counterparty
}

// GetConnectionHops implements Channel interface.
func (ch Channel) GetConnectionHops() []string {
	return ch.ConnectionHops
}

// GetVersion implements Channel interface.
func (ch Channel) GetVersion() string {
	return ch.Version
}

// ValidateBasic performs a basic validation of the channel fields
func (ch Channel) ValidateBasic() error {
	if ch.State == UNINITIALIZED {
		return ErrInvalidChannelState
	}
	if !(ch.Ordering == ORDERED || ch.Ordering == UNORDERED) {
		return sdkerrors.Wrap(ErrInvalidChannelOrdering, ch.Ordering.String())
	}
	if len(ch.ConnectionHops) != 1 {
		return sdkerrors.Wrap(
			ErrTooManyConnectionHops,
			"current IBC version only supports one connection hop",
		)
	}
	if err := host.ConnectionIdentifierValidator(ch.ConnectionHops[0]); err != nil {
		return sdkerrors.Wrap(err, "invalid connection hop ID")
	}
	return ch.Counterparty.ValidateBasic()
}

// NewCounterparty returns a new Counterparty instance
func NewCounterparty(portID, channelID string) Counterparty {
	return Counterparty{
		PortId:    portID,
		ChannelId: channelID,
	}
}

// GetPortID implements CounterpartyChannelI interface
func (c Counterparty) GetPortID() string {
	return c.PortId
}

// GetChannelID implements CounterpartyChannelI interface
func (c Counterparty) GetChannelID() string {
	return c.ChannelId
}

// ValidateBasic performs a basic validation check of the identifiers
func (c Counterparty) ValidateBasic() error {
	if err := host.PortIdentifierValidator(c.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid counterparty port ID")
	}
	if c.ChannelId != "" {
		if err := host.ChannelIdentifierValidator(c.ChannelId); err != nil {
			return sdkerrors.Wrap(err, "invalid counterparty channel ID")
		}
	}
	return nil
}

// NewIdentifiedChannel creates a new IdentifiedChannel instance
func NewIdentifiedChannel(portID, channelID string, ch Channel) IdentifiedChannel {
	return IdentifiedChannel{
		State:          ch.State,
		Ordering:       ch.Ordering,
		Counterparty:   ch.Counterparty,
		ConnectionHops: ch.ConnectionHops,
		Version:        ch.Version,
		PortId:         portID,
		ChannelId:      channelID,
	}
}

// ValidateBasic performs a basic validation of the identifiers and channel fields.
func (ic IdentifiedChannel) ValidateBasic() error {
	if err := host.ChannelIdentifierValidator(ic.ChannelId); err != nil {
		return sdkerrors.Wrap(err, "invalid channel ID")
	}
	if err := host.PortIdentifierValidator(ic.PortId); err != nil {
		return sdkerrors.Wrap(err, "invalid port ID")
	}
	channel := NewChannel(ic.State, ic.Ordering, ic.Counterparty, ic.ConnectionHops, ic.Version)
	return channel.ValidateBasic()
}

// NewResultAcknowledgement returns a new instance of Acknowledgement using an Acknowledgement_Result
// type in the Response field.
func NewResultAcknowledgement(result []byte) Acknowledgement {
	return Acknowledgement{
		Response: &Acknowledgement_Result{
			Result: result,
		},
	}
}

// NewErrorAcknowledgement returns a new instance of Acknowledgement using an Acknowledgement_Error
// type in the Response field.
func NewErrorAcknowledgement(err string) Acknowledgement {
	return Acknowledgement{
		Response: &Acknowledgement_Error{
			Error: err,
		},
	}
}

// GetBytes is a helper for serialising acknowledgements
func (ack Acknowledgement) GetBytes() []byte {
	return sdk.MustSortJSON(SubModuleCdc.MustMarshalJSON(&ack))
}

// ValidateBasic performs a basic validation of the acknowledgement
func (ack Acknowledgement) ValidateBasic() error {
	switch resp := ack.Response.(type) {
	case *Acknowledgement_Result:
		if len(resp.Result) == 0 {
			return sdkerrors.Wrap(ErrInvalidAcknowledgement, "acknowledgement result cannot be empty")
		}
	case *Acknowledgement_Error:
		if strings.TrimSpace(resp.Error) == "" {
			return sdkerrors.Wrap(ErrInvalidAcknowledgement, "acknowledgement error cannot be empty")
		}
	default:
		return sdkerrors.Wrapf(ErrInvalidAcknowledgement, "unsupported acknowledgement response field type %T", resp)
	}
	return nil
}
