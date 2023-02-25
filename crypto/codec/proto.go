package codec

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	var pub *cryptotypes.PubKey
	registry.RegisterInterface("cosmos.crypto.PubKey", pub)
	registry.RegisterImplementations(pub, &ed25519.PubKey{})
	registry.RegisterImplementations(pub, &secp256k1.PubKey{})
	registry.RegisterImplementations(pub, &secp256k1.EthPubKey{})
	registry.RegisterImplementations(pub, &multisig.LegacyAminoPubKey{})

	var priv *cryptotypes.PrivKey
	registry.RegisterInterface("cosmos.crypto.PrivKey", priv)
	registry.RegisterImplementations(priv, &secp256k1.PrivKey{})
	registry.RegisterImplementations(priv, &secp256k1.EthPrivKey{})
	registry.RegisterImplementations(priv, &ed25519.PrivKey{}) //nolint

	secp256r1.RegisterInterfaces(registry)
}
