package codec

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	// bls12_381 "github.com/cosmos/cosmos-sdk/crypto/keys/bls12_381"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1eth"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256r1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// RegisterInterfaces registers the crypto key interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	var pk *cryptotypes.PubKey
	registry.RegisterInterface("cosmos.crypto.PubKey", pk)
	registry.RegisterImplementations(pk, &ed25519.PubKey{})
	registry.RegisterImplementations(pk, &secp256k1.PubKey{})
	// registry.RegisterImplementations(pk, &bls12_381.PubKey{})
	registry.RegisterImplementations(pk, &mldsa65.PubKey{})
	registry.RegisterImplementations(pk, &secp256k1eth.PubKey{})
	registry.RegisterImplementations(pk, &multisig.LegacyAminoPubKey{})

	var priv *cryptotypes.PrivKey
	registry.RegisterInterface("cosmos.crypto.PrivKey", priv)
	registry.RegisterImplementations(priv, &secp256k1.PrivKey{})
	registry.RegisterImplementations(priv, &ed25519.PrivKey{})
	// registry.RegisterImplementations(priv, &bls12_381.PrivKey{})
	registry.RegisterImplementations(priv, &mldsa65.PrivKey{})
	registry.RegisterImplementations(priv, &secp256k1eth.PrivKey{})
	secp256r1.RegisterInterfaces(registry)
}
