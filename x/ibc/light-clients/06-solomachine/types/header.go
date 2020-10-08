package types

import (
	"strings"

	tmcrypto "github.com/tendermint/tendermint/crypto"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var _ exported.Header = Header{}

// ClientType defines that the Header is a Solo Machine.
func (Header) ClientType() string {
	return SoloMachine
}

// GetHeight returns the current sequence number as the height.
// Return clientexported.Height to satisfy interface
// Version number is always 0 for a solo-machine
func (h Header) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, h.Sequence)
}

// GetPubKey unmarshals the new public key into a tmcrypto.PubKey type.
func (h Header) GetPubKey() tmcrypto.PubKey {
	publicKey, ok := h.NewPublicKey.GetCachedValue().(tmcrypto.PubKey)
	if !ok {
		panic("Header NewPublicKey is not crypto.PubKey")
	}

	return publicKey
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

	if h.NewPublicKey == nil || h.GetPubKey() == nil || len(h.GetPubKey().Bytes()) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "new public key cannot be empty")
	}

	return nil
}
