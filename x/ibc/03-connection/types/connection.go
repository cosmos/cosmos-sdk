package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ exported.ConnectionI = (*ConnectionEnd)(nil)

// NewConnectionEnd creates a new ConnectionEnd instance.
func NewConnectionEnd(state State, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		ClientID:     clientID,
		Versions:     versions,
		State:        state,
		Counterparty: counterparty,
	}
}

// GetState implements the Connection interface
func (c ConnectionEnd) GetState() int32 {
	return int32(c.State)
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
	if err := host.ClientIdentifierValidator(c.ClientID); err != nil {
		return sdkerrors.Wrap(err, "invalid client ID")
	}
	if len(c.Versions) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "empty connection versions")
	}
	for _, version := range c.Versions {
		if err := ValidateVersion(version); err != nil {
			return err
		}
	}
	return c.Counterparty.ValidateBasic()
}

var _ exported.CounterpartyI = (*Counterparty)(nil)

// NewCounterparty creates a new Counterparty instance.
func NewCounterparty(clientID, connectionID string, prefix commitmenttypes.MerklePrefix) Counterparty {
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
	return &c.Prefix
}

// ValidateBasic performs a basic validation check of the identifiers and prefix
func (c Counterparty) ValidateBasic() error {
	if err := host.ConnectionIdentifierValidator(c.ConnectionID); err != nil {
		return sdkerrors.Wrap(err, "invalid counterparty connection ID")
	}
	if err := host.ClientIdentifierValidator(c.ClientID); err != nil {
		return sdkerrors.Wrap(err, "invalid counterparty client ID")
	}
	if c.Prefix.Empty() {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "counterparty prefix cannot be empty")
	}
	return nil
}

// NewIdentifiedConnection creates a new IdentifiedConnection instance
func NewIdentifiedConnection(connectionID string, conn ConnectionEnd) IdentifiedConnection {
	return IdentifiedConnection{
		ID:           connectionID,
		ClientID:     conn.ClientID,
		Versions:     conn.Versions,
		State:        conn.State,
		Counterparty: conn.Counterparty,
	}
}

// ValidateBasic performs a basic validation of the connection identifier and connection fields.
func (ic IdentifiedConnection) ValidateBasic() error {
	if err := host.ConnectionIdentifierValidator(ic.ID); err != nil {
		return sdkerrors.Wrap(err, "invalid connection ID")
	}
	connection := NewConnectionEnd(ic.State, ic.ClientID, ic.Counterparty, ic.Versions)
	return connection.ValidateBasic()
}
