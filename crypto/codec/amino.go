package codec

import (
	"github.com/cosmos/crypto/curves/sr25519"

	"cosmossdk.io/core/legacy"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc legacy.Amino) {
	cdc.RegisterInterface((*cryptotypes.PubKey)(nil), nil)
	cdc.RegisterConcrete(sr25519.PubKey{},
		sr25519.PubKeyName)
	cdc.RegisterConcrete(&ed25519.PubKey{},
		ed25519.PubKeyName)
	cdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName)
	cdc.RegisterConcrete(&kmultisig.LegacyAminoPubKey{},
		kmultisig.PubKeyAminoRoute)

	cdc.RegisterInterface((*cryptotypes.PrivKey)(nil), nil)
	cdc.RegisterConcrete(sr25519.PrivKey{},
		sr25519.PrivKeyName)
	cdc.RegisterConcrete(&ed25519.PrivKey{},
		ed25519.PrivKeyName)
	cdc.RegisterConcrete(&secp256k1.PrivKey{},
		secp256k1.PrivKeyName)
}
