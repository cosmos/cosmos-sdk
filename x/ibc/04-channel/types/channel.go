package types

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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
func (ch Channel) ValidateBasic() sdk.Error {
	if ch.State.String() == "" {
		return ErrInvalidChannelState(
			DefaultCodespace,
			"channel order should be either 'ORDERED' or 'UNORDERED'",
		)
	}
	if ch.Ordering.String() == "" {
		return ErrInvalidChannel(
			DefaultCodespace,
			"channel order should be either 'ORDERED' or 'UNORDERED'",
		)
	}
	if len(ch.ConnectionHops) != 1 {
		return ErrInvalidChannel(DefaultCodespace, "IBC v1 only supports one connection hop")
	}
	if err := host.DefaultConnectionIdentifierValidator(ch.ConnectionHops[0]); err != nil {
		return ErrInvalidChannel(DefaultCodespace, errors.Wrap(err, "invalid connection hop ID").Error())
	}
	if strings.TrimSpace(ch.Version) == "" {
		return ErrInvalidChannel(DefaultCodespace, "channel version can't be blank")
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
func (c Counterparty) ValidateBasic() sdk.Error {
	if err := host.DefaultPortIdentifierValidator(c.PortID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid counterparty connection ID: %s", err.Error()))
	}
	if err := host.DefaultChannelIdentifierValidator(c.ChannelID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid counterparty client ID: %s", err.Error()))
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
	OrderNone      string = "NONE"
	OrderUnordered string = "UNORDERED"
	OrderOrdered   string = "ORDERED"
)

// String implements the Stringer interface
func (o Order) String() string {
	switch o {
	case NONE:
		return OrderNone
	case UNORDERED:
		return OrderUnordered
	case ORDERED:
		return OrderOrdered
	default:
		return ""
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

	bz2, err := OrderFromString(s)
	if err != nil {
		return err
	}

	*o = bz2
	return nil
}

// OrderFromString parses a string into a channel order byte
func OrderFromString(order string) (Order, error) {
	switch order {
	case OrderNone:
		return NONE, nil
	case OrderUnordered:
		return UNORDERED, nil
	case OrderOrdered:
		return ORDERED, nil
	default:
		return 0, fmt.Errorf("'%s' is not a valid channel ordering", order)
	}
}

// State defines if a channel is in one of the following states:
// CLOSED, INIT, OPENTRY or OPEN
type State byte

// channel state types
const (
	CLOSED  State = iota + 1 // A channel end has been closed and can no longer be used to send or receive packets.
	INIT                     // A channel end has just started the opening handshake.
	OPENTRY                  // A channel end has acknowledged the handshake step on the counterparty chain.
	OPEN                     // A channel end has completed the handshake and is ready to send and receive packets.
)

// string representation of the channel states
const (
	StateClosed  string = "CLOSED"
	StateInit    string = "INIT"
	StateOpenTry string = "OPENTRY"
	StateOpen    string = "OPEN"
)

// String implements the Stringer interface
func (s State) String() string {
	switch s {
	case CLOSED:
		return StateClosed
	case INIT:
		return StateInit
	case OPENTRY:
		return StateOpenTry
	case OPEN:
		return StateOpen
	default:
		return ""
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

	bz2, err := StateFromString(stateStr)
	if err != nil {
		return err
	}

	*s = bz2
	return nil
}

// StateFromString parses a string into a channel state byte
func StateFromString(state string) (State, error) {
	switch state {
	case StateClosed:
		return CLOSED, nil
	case StateInit:
		return INIT, nil
	case StateOpenTry:
		return OPENTRY, nil
	case StateOpen:
		return OPEN, nil
	default:
		return CLOSED, fmt.Errorf("'%s' is not a valid channel state", state)
	}
}
