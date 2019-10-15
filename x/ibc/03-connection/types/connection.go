package types

import (
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ICS03 - Connection Data Structures as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#data-structures

// ConnectionEnd defines a stateful object on a chain connected to another separate
// one.
// NOTE: there must only be 2 defined ConnectionEnds to stablish a connection
// between two chains.
type ConnectionEnd struct {
	State    ConnectionState `json:"state" yaml:"state"`
	ClientID string          `json:"client_id" yaml:"client_id"`

	// Counterparty chain associated with this connection.
	Counterparty Counterparty `json:"counterparty" yaml:"counterparty"`
	// Version is utilised to determine encodings or protocols for channels or
	// packets utilising this connection.
	Versions []string `json:"versions" yaml:"versions"`
}

// NewConnectionEnd creates a new ConnectionEnd instance.
func NewConnectionEnd(state ConnectionState, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     versions,
	}
}

// LatestVersion gets the latest version of a connection protocol
func (ce ConnectionEnd) LatestVersion() string {
	if len(ce.Versions) == 0 {
		return ""
	}
	return ce.Versions[len(ce.Versions)-1]
}

// TODO: create a custom JSON marshaler

// Counterparty defines the counterparty chain associated with a connection end.
type Counterparty struct {
	ClientID     string       `json:"client_id" yaml:"client_id"`
	ConnectionID string       `json:"connection_id" yaml:"connection_id"`
	Prefix       ics23.Prefix `json:"prefix" yaml:"prefix`
}

// NewCounterparty creates a new Counterparty instance.
func NewCounterparty(clientID, connectionID string, prefix ics23.Prefix) Counterparty {
	return Counterparty{
		ClientID:     clientID,
		ConnectionID: connectionID,
		Prefix:       prefix,
	}
}

// ConnectionState defines the state of a connection between two disctinct
// chains
type ConnectionState = byte

// available connection states
const (
	NONE ConnectionState = iota // default ConnectionState
	INIT
	TRYOPEN
	OPEN
)

// ConnectionStateToString returns the string representation of a connection state
func ConnectionStateToString(state ConnectionState) string {
	switch state {
	case NONE:
		return "NONE"
	case INIT:
		return "INIT"
	case TRYOPEN:
		return "TRYOPEN"
	case OPEN:
		return "OPEN"
	default:
		return ""
	}
}

// StringToConnectionState parses a string into a connection state
func StringToConnectionState(state string) ConnectionState {
	switch state {
	case "NONE":
		return NONE
	case "INIT":
		return INIT
	case "TRYOPEN":
		return TRYOPEN
	case "OPEN":
		return OPEN
	default:
		return NONE
	}
}
