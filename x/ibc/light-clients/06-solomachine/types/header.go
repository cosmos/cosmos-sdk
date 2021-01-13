package types

import (
	"strings"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var _ exported.Header = &Header{}

// ClientType defines that the Header is a Solo Machine.
func (Header) ClientType() string {
	return exported.Solomachine
}

// GetHeight returns the current sequence number as the height.
// Return clientexported.Height to satisfy interface
// Revision number is always 0 for a solo-machine
func (h Header) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, h.Sequence)
}

// GetPubKey unmarshals the new public key into a cryptotypes.PubKey type.
// An error is returned if the new public key is nil or the cached value
// is not a PubKey.
func (h Header) GetPubKey() (cryptotypes.PubKey, error) {
	if h.NewPublicKey == nil {
		return nil, sdkerrors.Wrap(ErrInvalidHeader, "header NewPublicKey cannot be nil")
	}

	publicKey, ok := h.NewPublicKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrap(ErrInvalidHeader, "header NewPublicKey is not cryptotypes.PubKey")
	}

	return publicKey, nil
}

// ValidateBasic ensures that the sequence, signature and public key have all
// been initialized.
func (h Header) ValidateBasic() error {
	if h.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "sequence number cannot be zero")
	}

	if h.Timestamp == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "timestamp cannot be zero")
	}

	if h.NewDiversifier != "" && strings.TrimSpace(h.NewDiversifier) == "" {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "diversifier cannot contain only spaces")
	}

	if len(h.Signature) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "signature cannot be empty")
	}

	newPublicKey, err := h.GetPubKey()
	if err != nil || newPublicKey == nil || len(newPublicKey.Bytes()) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "new public key cannot be empty")
	}

	return nil
}
