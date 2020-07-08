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
func NewConnectionEnd(state State, connectionID, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		ID:           connectionID,
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
	if err := host.ConnectionIdentifierValidator(c.ID); err != nil {
		return sdkerrors.Wrapf(err, "invalid connection ID: %s", c.ID)
	}
	if err := host.ClientIdentifierValidator(c.ClientID); err != nil {
		return sdkerrors.Wrapf(err, "invalid client ID: %s", c.ClientID)
	}
	if len(c.Versions) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidVersion, "missing connection versions")
	}
	for _, version := range c.Versions {
		if err := host.ConnectionVersionValidator(version); err != nil {
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
