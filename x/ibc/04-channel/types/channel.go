package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var (
	_ exported.ChannelI      = (*Channel)(nil)
	_ exported.CounterpartyI = (*Counterparty)(nil)
)

// NewChannel creates a new Channel instance
func NewChannel(
	state ibctypes.State, ordering ibctypes.Order, counterparty Counterparty,
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
func (ch Channel) GetState() ibctypes.State {
	return ch.State
}

// GetOrdering implements Channel interface.
func (ch Channel) GetOrdering() ibctypes.Order {
	return ch.Ordering
}

// GetCounterparty implements Channel interface.
func (ch Channel) GetCounterparty() exported.CounterpartyI {
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
	if ch.State.String() == "" {
		return sdkerrors.Wrap(ErrInvalidChannel, ErrInvalidChannelState.Error())
	}
	if !(ch.Ordering == ibctypes.ORDERED || ch.Ordering == ibctypes.UNORDERED) {
		return sdkerrors.Wrap(ErrInvalidChannelOrdering, ch.Ordering.String())
	}
	if len(ch.ConnectionHops) != 1 {
		return sdkerrors.Wrap(
			ErrInvalidChannel,
			sdkerrors.Wrap(ErrTooManyConnectionHops, "IBC v1.0 only supports one connection hop").Error(),
		)
	}
	if err := host.DefaultConnectionIdentifierValidator(ch.ConnectionHops[0]); err != nil {
		return sdkerrors.Wrap(
			ErrInvalidChannel,
			sdkerrors.Wrap(err, "invalid connection hop ID").Error(),
		)
	}
	if strings.TrimSpace(ch.Version) == "" {
		return sdkerrors.Wrap(
			ErrInvalidChannel,
			sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "channel version can't be blank").Error(),
		)
	}
	return ch.Counterparty.ValidateBasic()
}

// NewCounterparty returns a new Counterparty instance
func NewCounterparty(portID, channelID string) Counterparty {
	return Counterparty{
		PortID:    portID,
		ChannelID: channelID,
	}
}

// GetPortID implements CounterpartyI interface
func (c Counterparty) GetPortID() string {
	return c.PortID
}

// GetChannelID implements CounterpartyI interface
func (c Counterparty) GetChannelID() string {
	return c.ChannelID
}

// ValidateBasic performs a basic validation check of the identifiers
func (c Counterparty) ValidateBasic() error {
	if err := host.DefaultPortIdentifierValidator(c.PortID); err != nil {
		return sdkerrors.Wrap(
			ErrInvalidCounterparty,
			sdkerrors.Wrap(err, "invalid counterparty connection ID").Error(),
		)
	}
	if err := host.DefaultChannelIdentifierValidator(c.ChannelID); err != nil {
		return sdkerrors.Wrap(
			ErrInvalidCounterparty,
			sdkerrors.Wrap(err, "invalid counterparty client ID").Error(),
		)
	}
	return nil
}

// IdentifiedChannel defines a channel with additional port and channel identifier
// fields.
type IdentifiedChannel struct {
	ID             string         `json:"id" yaml:"id"`
	PortID         string         `json:"port_id" yaml:"port_id"`
	State          ibctypes.State `json:"state" yaml:"state"`
	Ordering       ibctypes.Order `json:"ordering" yaml:"ordering"`
	Counterparty   Counterparty   `json:"counterparty" yaml:"counterparty"`
	ConnectionHops []string       `json:"connection_hops" yaml:"connection_hops"`
	Version        string         `json:"version" yaml:"version "`
}

// NewIdentifiedChannel creates a new IdentifiedChannel instance
func NewIdentifiedChannel(portID, channelID string, ch Channel) IdentifiedChannel {
	return IdentifiedChannel{
		ID:             channelID,
		PortID:         portID,
		State:          ch.State,
		Ordering:       ch.Ordering,
		Counterparty:   ch.Counterparty,
		ConnectionHops: ch.ConnectionHops,
		Version:        ch.Version,
	}
}

// ValidateBasic performs a basic validation of the identifiers and channel fields.
func (ic IdentifiedChannel) ValidateBasic() error {
	if err := host.DefaultChannelIdentifierValidator(ic.ID); err != nil {
		return sdkerrors.Wrap(ErrInvalidChannel, err.Error())
	}
	if err := host.DefaultPortIdentifierValidator(ic.PortID); err != nil {
		return sdkerrors.Wrap(ErrInvalidChannel, err.Error())
	}
	channel := NewChannel(ic.State, ic.Ordering, ic.Counterparty, ic.ConnectionHops, ic.Version)
	return channel.ValidateBasic()
}
