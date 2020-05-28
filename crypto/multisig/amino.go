package multisig

import (
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto/multisig"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

var cdc = amino.NewCodec()

func init() {
	RegisterAmino(cdc)
}

func RegisterAmino(amino *amino.Codec) {
	amino.RegisterInterface((*crypto.PubKey)(nil), nil)
	amino.RegisterConcrete(ThresholdMultisigPubKey{},
		multisig.PubKeyMultisigThresholdAminoRoute, nil)
	amino.RegisterConcrete(ed25519.PubKeyEd25519{},
		ed25519.PubKeyAminoName, nil)
	amino.RegisterConcrete(sr25519.PubKeySr25519{},
		sr25519.PubKeyAminoName, nil)
	amino.RegisterConcrete(secp256k1.PubKeySecp256k1{},
		secp256k1.PubKeyAminoName, nil)
}
