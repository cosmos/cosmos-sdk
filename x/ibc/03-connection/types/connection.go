package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmentexported "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ exported.ConnectionI = (*ConnectionEnd)(nil)

// NewConnectionEnd creates a new ConnectionEnd instance.
func NewConnectionEnd(state ibctypes.State, connectionID, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		ID:           connectionID,
		ClientID:     clientID,
		Versions:     versions,
		State:        state,
		Counterparty: counterparty,
	}
}

// GetState implements the Connection interface
func (c ConnectionEnd) GetState() ibctypes.State {
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
	if c.Prefix.IsEmpty() {
		return sdkerrors.Wrap(ErrInvalidCounterparty, "invalid counterparty prefix")
	}
	return nil
}
