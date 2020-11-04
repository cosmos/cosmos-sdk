package codec

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmprotocrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// FromTmPubKey converts TM's tmcrypto.PubKey toour own PubKey.
func FromTmPubKey(tmPk tmcrypto.PubKey) (cryptotypes.PubKey, error) {
	switch tmPk := tmPk.(type) {
	case tmed25519.PubKey:
		return &ed25519.PubKey{Key: []byte(tmPk)}, nil
	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to cryptotypes.PubKey", tmPk)
	}

}

// ToTmPubKey converts our own PubKey to TM's tmcrypto.PubKey.
func ToTmPubKey(pk cryptotypes.PubKey) (tmcrypto.PubKey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return tmed25519.PubKey(pk.Key), nil
	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to Tendermint public key", pk)
	}

}

// ToTmPublicKey converts our own PubKey to TM's tmprotocrypto.PublicKey.
func ToTmPublicKey(pk cryptotypes.PubKey) (tmprotocrypto.PublicKey, error) {
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		return tmprotocrypto.PublicKey{
			Sum: &tmprotocrypto.PublicKey_Ed25519{
				Ed25519: pk.Key,
			},
		}, nil
	default:
		return tmprotocrypto.PublicKey{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to Tendermint public key", pk)
	}
}
