package codec

import (
	protocrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface("cosmos.crypto.PubKey", (*cryptotypes.PubKey)(nil))
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &ed25519.PubKey{})
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &secp256k1.PubKey{})
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &multisig.LegacyAminoPubKey{})
}

// ToTmPubKey converts our own PubKey to TM's protocrypto.PublicKey.
func ToTmPubKey(pk cryptotypes.PubKey) (protocrypto.PublicKey, error) {
	var tmPk protocrypto.PublicKey
	switch pk := pk.(type) {
	case *ed25519.PubKey:
		tmPk = protocrypto.PublicKey{
			Sum: &protocrypto.PublicKey_Ed25519{
				Ed25519: pk.Key,
			},
		}
	default:
		return tmPk, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "cannot convert %v to Tendermint public key", pk)
	}

	return tmPk, nil
}
