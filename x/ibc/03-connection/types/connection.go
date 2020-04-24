package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ exported.ConnectionI = ConnectionEnd{}

// ICS03 - Connection Data Structures as defined in https://github.com/cosmos/ics/tree/master/spec/ics-003-connection-semantics#data-structures

// ConnectionEnd defines a stateful object on a chain connected to another separate
// one.
// NOTE: there must only be 2 defined ConnectionEnds to establish a connection
// between two chains.
type ConnectionEnd struct {
	State    exported.State `json:"state" yaml:"state"`
	ID       string         `json:"id" yaml:"id"`
	ClientID string         `json:"client_id" yaml:"client_id"`

	// Counterparty chain associated with this connection.
	Counterparty Counterparty `json:"counterparty" yaml:"counterparty"`
	// Version is utilised to determine encodings or protocols for channels or
	// packets utilising this connection.
	Versions []string `json:"versions" yaml:"versions"`
}

// NewConnectionEnd creates a new ConnectionEnd instance.
func NewConnectionEnd(state exported.State, connectionID, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		State:        state,
		ID:           connectionID,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     versions,
	}
}

// GetState implements the Connection interface
func (c ConnectionEnd) GetState() exported.State {
	return c.State
}

// GetID implements the Connection interface
func (c ConnectionEnd) GetID() string {
	return c.ID
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

// ValidateBasic implements the Connection interface.
// NOTE: the protocol supports that the connection and client IDs match the
// counterparty's.
func (c ConnectionEnd) ValidateBasic() error {
	if err := host.DefaultConnectionIdentifierValidator(c.ID); err != nil {
		return sdkerrors.Wrapf(err, "invalid connection ID: %s", c.ID)
	}
	if err := host.DefaultClientIdentifierValidator(c.ClientID); err != nil {
		return sdkerrors.Wrapf(err, "invalid client ID: %s", c.ClientID)
	}
	if len(c.Versions) == 0 {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "missing connection versions")
	}
	for _, version := range c.Versions {
		if strings.TrimSpace(version) == "" {
			return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "version can't be blank")
		}
	}
	return c.Counterparty.ValidateBasic()
}

var _ exported.CounterpartyI = Counterparty{}

// Counterparty defines the counterparty chain associated with a connection end.
type Counterparty struct {
	ClientID     string                    `json:"client_id" yaml:"client_id"`
	ConnectionID string                    `json:"connection_id" yaml:"connection_id"`
	Prefix       commitmentexported.Prefix `json:"prefix" yaml:"prefix"`
}

// NewCounterparty creates a new Counterparty instance.
func NewCounterparty(clientID, connectionID string, prefix commitmentexported.Prefix) Counterparty {
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

// GetPrefix implements the CounterpartyI interface
func (c Counterparty) GetPrefix() commitmentexported.Prefix {
	return c.Prefix
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
