package types

import (
	"encoding/json"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	exported "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ exported.ConnectionI = ConnectionEnd{}

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

// GetState implements the Connection interface
func (c ConnectionEnd) GetState() exported.StateI {
	return &c.State
}

// GetClientID implements the Connection interface
func (c ConnectionEnd) GetClientID() string {
	return c.ClientID
}

// GetCounterparty implements the Connection interface
func (c ConnectionEnd) GetCounterparty() exported.CounterpartyI {
	return c.Counterparty
}

// GetVersions implements the Connection interface
func (c ConnectionEnd) GetVersions() []string {
	return c.Versions
}

// ValidateBasic implements the Connection interface
func (c ConnectionEnd) ValidateBasic() error {
	// TODO:
	return c.Counterparty.ValidateBasic()
}

var _ exported.CounterpartyI = Counterparty{}

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

// GetClientID implements the CounterpartyI interface
func (c Counterparty) GetClientID() string {
	return c.ClientID
}

// GetConnectionID implements the CounterpartyI interface
func (c Counterparty) GetConnectionID() string {
	return c.ConnectionID
}

// ValidateBasic performs a basic validation check of the identifiers and prefix
func (c Counterparty) ValidateBasic() error {
	if err := host.DefaultConnectionIdentifierValidator(c.ConnectionID); err != nil {
		return sdkerrors.Wrap(err,
			sdkerrors.Wrapf(
				ErrInvalidCounterparty,
				"invalid counterparty connection ID %s", c.ConnectionID,
			).Error(),
		)
	}
	if err := host.DefaultClientIdentifierValidator(c.ClientID); err != nil {
		return sdkerrors.Wrap(err,
			sdkerrors.Wrapf(
				ErrInvalidCounterparty,
				"invalid counterparty client ID %s", c.ClientID,
			).Error(),
		)
	}
	if c.Prefix == nil || len(c.Prefix.Bytes()) == 0 {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "invalid counterparty prefix")
	}
	return nil
}

var s = State(0)
var _ exported.StateI = &s

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
