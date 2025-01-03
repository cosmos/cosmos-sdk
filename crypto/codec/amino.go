package codec

import (
	"github.com/cometbft/cometbft/crypto/bls12381"

	"cosmossdk.io/core/registry"

	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12_381"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(registrar registry.AminoRegistrar) {
	registrar.RegisterInterface((*cryptotypes.PubKey)(nil), nil)
	registrar.RegisterConcrete(&ed25519.PubKey{},
		ed25519.PubKeyName)
	registrar.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName)
	registrar.RegisterConcrete(&bls12_381.PubKey{}, bls12381.PubKeyName)
	registrar.RegisterConcrete(&kmultisig.LegacyAminoPubKey{},
		kmultisig.PubKeyAminoRoute)
	registrar.RegisterInterface((*cryptotypes.PrivKey)(nil), nil)
	registrar.RegisterConcrete(&ed25519.PrivKey{},
		ed25519.PrivKeyName)
	registrar.RegisterConcrete(&secp256k1.PrivKey{},
		secp256k1.PrivKeyName)
	registrar.RegisterConcrete(&bls12_381.PrivKey{}, bls12381.PrivKeyName)
}
