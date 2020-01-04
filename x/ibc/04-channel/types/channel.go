package types

import (
	"encoding/json"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

type Channel struct {
	State          State        `json:"state" yaml:"state"`
	Ordering       Order        `json:"ordering" yaml:"ordering"`
	Counterparty   Counterparty `json:"counterparty" yaml:"counterparty"`
	ConnectionHops []string     `json:"connection_hops" yaml:"connection_hops"`
	Version        string       `json:"version" yaml:"version "`
}

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

// ValidateBasic performs a basic validation of the channel fields
func (ch Channel) ValidateBasic() error {
	if ch.State.String() == "" {
		return sdkerrors.Wrap(ErrInvalidChannel, ErrInvalidChannelState.Error())
	}
	if ch.Ordering.String() == "" {
		return sdkerrors.Wrap(ErrInvalidChannel, ErrInvalidChannelOrdering.Error())
	}
	if len(ch.ConnectionHops) != 1 {
		return sdkerrors.Wrap(ErrInvalidChannel, "IBC v1 only supports one connection hop")
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

// Counterparty defines the counterparty chain's channel and port identifiers
type Counterparty struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
}

// NewCounterparty returns a new Counterparty instance
func NewCounterparty(portID, channelID string) Counterparty {
	return Counterparty{
		PortID:    portID,
		ChannelID: channelID,
	}
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

// Order defines if a channel is ORDERED or UNORDERED
type Order byte

// string representation of the channel ordering
const (
	NONE      Order = iota // zero-value for channel ordering
	UNORDERED              // packets can be delivered in any order, which may differ from the order in which they were sent.
	ORDERED                // packets are delivered exactly in the order which they were sent
)

// channel order types
const (
	OrderNone      string = ""
	OrderUnordered string = "UNORDERED"
	OrderOrdered   string = "ORDERED"
)

// String implements the Stringer interface
func (o Order) String() string {
	switch o {
	case UNORDERED:
		return OrderUnordered
	case ORDERED:
		return OrderOrdered
	default:
		return OrderNone
	}
}

// MarshalJSON marshal to JSON using string.
func (o Order) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

// UnmarshalJSON decodes from JSON.
func (o *Order) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	order := OrderFromString(s)
	if order == 0 {
		return sdkerrors.Wrap(ErrInvalidChannelOrdering, s)
	}

	*o = order
	return nil
}

// OrderFromString parses a string into a channel order byte
func OrderFromString(order string) Order {
	switch order {
	case OrderUnordered:
		return UNORDERED
	case OrderOrdered:
		return ORDERED
	default:
		return NONE
	}
}

// State defines if a channel is in one of the following states:
// CLOSED, INIT, OPENTRY or OPEN
type State byte

// channel state types
const (
	UNINITIALIZED State = iota // Default State
	INIT                       // A channel end has just started the opening handshake.
	TRYOPEN                    // A channel end has acknowledged the handshake step on the counterparty chain.
	OPEN                       // A channel end has completed the handshake and is ready to send and receive packets.
	CLOSED                     // A channel end has been closed and can no longer be used to send or receive packets.
)

// string representation of the channel states
const (
	StateUninitialized string = "UNINITIALIZED"
	StateInit          string = "INIT"
	StateTryOpen       string = "TRYOPEN"
	StateOpen          string = "OPEN"
	StateClosed        string = "CLOSED"
)

// String implements the Stringer interface
func (s State) String() string {
	switch s {
	case INIT:
		return StateInit
	case TRYOPEN:
		return StateTryOpen
	case OPEN:
		return StateOpen
	case CLOSED:
		return StateClosed
	default:
		return StateUninitialized
	}
}

// MarshalJSON marshal to JSON using string.
func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON decodes from JSON.
func (s *State) UnmarshalJSON(data []byte) error {
	var stateStr string
	err := json.Unmarshal(data, &stateStr)
	if err != nil {
		return err
	}

	*s = StateFromString(stateStr)
	return nil
}

// StateFromString parses a string into a channel state byte
func StateFromString(state string) State {
	switch state {
	case StateClosed:
		return CLOSED
	case StateInit:
		return INIT
	case StateTryOpen:
		return TRYOPEN
	case StateOpen:
		return OPEN
	default:
		return UNINITIALIZED
	}
}
