package types

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// ICS03 - Connection Data Structures as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#data-structures

// ConnectionEnd defines a stateful object on a chain connected to another separate
// one.
// NOTE: there must only be 2 defined ConnectionEnds to stablish a connection
// between two chains.
type ConnectionEnd struct {
	State    State  `json:"state" yaml:"state"`
	ClientID string `json:"client_id" yaml:"client_id"`

	// Counterparty chain associated with this connection.
	Counterparty Counterparty `json:"counterparty" yaml:"counterparty"`
	// Version is utilised to determine encodings or protocols for channels or
	// packets utilising this connection.
	Versions []string `json:"versions" yaml:"versions"`
}

// NewConnectionEnd creates a new ConnectionEnd instance.
func NewConnectionEnd(state State, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     versions,
	}
}

// TODO: create a custom JSON marshaler

// Counterparty defines the counterparty chain associated with a connection end.
type Counterparty struct {
	ClientID     string             `json:"client_id" yaml:"client_id"`
	ConnectionID string             `json:"connection_id" yaml:"connection_id"`
	Prefix       commitment.PrefixI `json:"prefix" yaml:"prefix"`
}

// NewCounterparty creates a new Counterparty instance.
func NewCounterparty(clientID, connectionID string, prefix commitment.PrefixI) Counterparty {
	return Counterparty{
		ClientID:     clientID,
		ConnectionID: connectionID,
		Prefix:       prefix,
	}
}

// ValidateBasic performs a basic validation check of the identifiers and prefix
func (c Counterparty) ValidateBasic() sdk.Error {
	if err := host.DefaultConnectionIdentifierValidator(c.ConnectionID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid counterparty connection ID: %s", err.Error()))
	}
	if err := host.DefaultClientIdentifierValidator(c.ClientID); err != nil {
		return sdk.NewError(host.IBCCodeSpace, 1, fmt.Sprintf("invalid counterparty client ID: %s", err.Error()))
	}
	if c.Prefix == nil || len(c.Prefix.Bytes()) == 0 {
		return ErrInvalidCounterparty(DefaultCodespace, "invalid counterparty prefix")
	}
	return nil
}

// State defines the state of a connection between two disctinct
// chains
type State byte

// available connection states
const (
	NONE State = iota // default State
	INIT
	TRYOPEN
	OPEN
)

// string representation of the connection states
const (
	StateNone    string = "NONE"
	StateInit    string = "INIT"
	StateTryOpen string = "TRYOPEN"
	StateOpen    string = "OPEN"
)

// String implements the Stringer interface
func (cs State) String() string {
	switch cs {
	case NONE:
		return StateNone
	case INIT:
		return StateInit
	case TRYOPEN:
		return StateTryOpen
	case OPEN:
		return StateOpen
	default:
		return ""
	}
}

// StateFromString parses a string into a connection state
func StateFromString(state string) (State, error) {
	switch state {
	case StateNone:
		return NONE, nil
	case StateInit:
		return INIT, nil
	case StateTryOpen:
		return TRYOPEN, nil
	case StateOpen:
		return OPEN, nil
	default:
		return NONE, fmt.Errorf("'%s' is not a valid connection state", state)
	}
}

// MarshalJSON marshal to JSON using string.
func (cs State) MarshalJSON() ([]byte, error) {
	return json.Marshal(cs.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (cs *State) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := StateFromString(s)
	if err != nil {
		return err
	}

	*cs = bz2
	return nil
}
