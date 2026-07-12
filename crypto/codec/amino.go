package codec

import (
	"github.com/cometbft/cometbft/crypto/bls12381"
	cmtmldsa65 "github.com/cometbft/cometbft/crypto/mldsa65"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12_381"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*cryptotypes.PubKey)(nil), nil)
	cdc.RegisterConcrete(&ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	cdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)
	cdc.RegisterConcrete(&kmultisig.LegacyAminoPubKey{},
		kmultisig.PubKeyAminoRoute, nil)
	cdc.RegisterConcrete(&bls12_381.PubKey{}, bls12381.PubKeyName, nil)
	cdc.RegisterConcrete(&mldsa65.PubKey{}, cmtmldsa65.PubKeyName, nil)

	cdc.RegisterInterface((*cryptotypes.PrivKey)(nil), nil)
	cdc.RegisterConcrete(&ed25519.PrivKey{},
		ed25519.PrivKeyName, nil)
	cdc.RegisterConcrete(&secp256k1.PrivKey{},
		secp256k1.PrivKeyName, nil)
	cdc.RegisterConcrete(&bls12_381.PrivKey{}, bls12381.PrivKeyName, nil)
	cdc.RegisterConcrete(&mldsa65.PrivKey{}, cmtmldsa65.PrivKeyName, nil)
}
