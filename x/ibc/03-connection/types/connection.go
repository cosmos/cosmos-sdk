package types

import (
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// ICS03 - Connection Data Structures as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#data-structures

// ConnectionState defines the state of a connection between two disctinct
// chains
type ConnectionState = byte

// available ConnectionStates
const (
	NONE ConnectionState = iota // default State
	INIT
	TRYOPEN
	OPEN
)

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
func NewConnectionEnd(clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		State:        NONE,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     versions,
	}
}

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
