package codec

import (
	codectypes "github.com/pointnetwork/cosmos-point-sdk/codec/types"
	"github.com/pointnetwork/cosmos-point-sdk/crypto/keys/ed25519"
	"github.com/pointnetwork/cosmos-point-sdk/crypto/keys/multisig"
	"github.com/pointnetwork/cosmos-point-sdk/crypto/keys/secp256k1"
	"github.com/pointnetwork/cosmos-point-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/pointnetwork/cosmos-point-sdk/crypto/types"
)

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	var pk *cryptotypes.PubKey
	registry.RegisterInterface("cosmos.crypto.PubKey", pk)
	registry.RegisterImplementations(pk, &ed25519.PubKey{})
	registry.RegisterImplementations(pk, &secp256k1.PubKey{})
	registry.RegisterImplementations(pk, &multisig.LegacyAminoPubKey{})

	var priv *cryptotypes.PrivKey
	registry.RegisterInterface("cosmos.crypto.PrivKey", priv)
	registry.RegisterImplementations(priv, &secp256k1.PrivKey{})
	registry.RegisterImplementations(priv, &ed25519.PrivKey{}) //nolint
	secp256r1.RegisterInterfaces(registry)
}
