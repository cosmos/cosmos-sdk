package exported

import (
	"encoding/json"

	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// ConnectionI describes the required methods for a connection.
type ConnectionI interface {
	GetState() State
	GetClientID() string
	GetCounterparty() CounterpartyI
	GetVersions() []string
	ValidateBasic() error
}

// CounterpartyI describes the required methods for a counterparty connection.
type CounterpartyI interface {
	GetClientID() string
	GetConnectionID() string
	GetPrefix() commitmentexported.Prefix
	ValidateBasic() error
}

// State defines the state of a connection between two disctinct
// chains
type State byte

// available connection states
const (
	UNINITIALIZED State = iota // default State
	INIT
	TRYOPEN
	OPEN
)

// string representation of the connection states
const (
	StateUninitialized string = "UNINITIALIZED"
	StateInit          string = "INIT"
	StateTryOpen       string = "TRYOPEN"
	StateOpen          string = "OPEN"
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
	default:
		return StateUninitialized
	}
}

// StateFromString parses a string into a connection state
func StateFromString(state string) State {
	switch state {
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

// MarshalJSON marshal to JSON using string.
func (s State) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (s *State) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}

	*s = StateFromString(str)
	return nil
}
