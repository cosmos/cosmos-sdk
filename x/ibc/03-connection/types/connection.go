package types

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

var _ exported.ConnectionI = (*ConnectionEnd)(nil)

// NewConnectionEnd creates a new ConnectionEnd instance.
func NewConnectionEnd(state State, clientID string, counterparty Counterparty, versions []string) ConnectionEnd {
	return ConnectionEnd{
		State:        state,
		ClientID:     clientID,
		Counterparty: counterparty,
		Versions:     versions,
	}
}

// ValidateBasic implements the Connection interface
func (c ConnectionEnd) ValidateBasic() error {
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
