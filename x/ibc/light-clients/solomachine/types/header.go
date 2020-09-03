package types

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/std"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

var _ exported.Header = Header{}

// ClientType defines that the Header is a Solo Machine.
func (Header) ClientType() exported.ClientType {
	return exported.SoloMachine
}

// GetHeight returns the current sequence number as the height.
func (h Header) GetHeight() uint64 {
	return h.Sequence
}

// GetPubKey unmarshals the new public key into a tmcrypto.PubKey type.
func (h Header) GetPubKey() tmcrypto.PubKey {
	publicKey, err := std.DefaultPublicKeyCodec{}.Decode(h.NewPublicKey)
	if err != nil {
		panic(err)
	}

	return publicKey
}

// ValidateBasic ensures that the sequence, signature and public key have all
// been initialized.
func (h Header) ValidateBasic() error {
	if h.Sequence == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "sequence number cannot be zero")
	}

	if len(h.Signature) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "signature cannot be empty")
	}

	if h.NewPublicKey == nil || h.GetPubKey() == nil || len(h.GetPubKey().Bytes()) == 0 {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "new public key cannot be empty")
	}

	return nil
}
