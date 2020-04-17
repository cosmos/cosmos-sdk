package exported

import (
	"encoding/json"
	"fmt"
)

// ChannelI defines the standard interface for a channel end.
type ChannelI interface {
	GetState() State
	GetOrdering() Order
	GetCounterparty() CounterpartyI
	GetConnectionHops() []string
	GetVersion() string
	ValidateBasic() error
}

// CounterpartyI defines the standard interface for a channel end's
// counterparty.
type CounterpartyI interface {
	GetPortID() string
	GetChannelID() string
	ValidateBasic() error
}

// PacketI defines the standard interface for IBC packets
type PacketI interface {
	GetSequence() uint64
	GetTimeoutHeight() uint64
	GetTimeoutTimestamp() uint64
	GetSourcePort() string
	GetSourceChannel() string
	GetDestPort() string
	GetDestChannel() string
	GetData() []byte
	ValidateBasic() error
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
		return fmt.Errorf("invalid order '%s'", s)
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
