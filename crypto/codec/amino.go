package codec

import (
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
)

var amino *codec.Codec

func init() {
	amino = codec.New()
	RegisterCrypto(amino)
}

// RegisterCrypto registers all crypto dependency types with the provided Amino
// codec.
func RegisterCrypto(cdc *codec.Codec) {
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	cdc.RegisterConcrete(sr25519.PubKey{},
		sr25519.PubKeyName, nil)
	cdc.RegisterConcrete(secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)
	cdc.RegisterConcrete(multisig.PubKeyMultisigThreshold{},
		multisig.PubKeyAminoRoute, nil)

	cdc.RegisterInterface((*crypto.PrivKey)(nil), nil)
	cdc.RegisterConcrete(ed25519.PrivKey{},
		ed25519.PrivKeyName, nil)
	cdc.RegisterConcrete(sr25519.PrivKey{},
		sr25519.PrivKeyName, nil)
	cdc.RegisterConcrete(secp256k1.PrivKey{},
		secp256k1.PrivKeyName, nil)
}

// PrivKeyFromBytes unmarshals private key bytes and returns a PrivKey
func PrivKeyFromBytes(privKeyBytes []byte) (privKey crypto.PrivKey, err error) {
	err = amino.UnmarshalBinaryBare(privKeyBytes, &privKey)
	return
}

// PubKeyFromBytes unmarshals public key bytes and returns a PubKey
func PubKeyFromBytes(pubKeyBytes []byte) (pubKey crypto.PubKey, err error) {
	err = amino.UnmarshalBinaryBare(pubKeyBytes, &pubKey)
	return
}
