package multisig

import (
	amino "github.com/tendermint/go-amino"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

// TODO: Figure out API for others to either add their own pubkey types, or
// to make verify / marshal accept a Cdc.
const (
	PubKeyAminoRoute = "tendermint/PubKeyMultisigThreshold"
)

var Cdc = amino.NewCodec()

func init() {
	Cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	Cdc.RegisterConcrete(PubKeyMultisigThreshold{},
		PubKeyAminoRoute, nil)
	Cdc.RegisterConcrete(ed25519.PubKeyEd25519{},
		ed25519.PubKeyAminoName, nil)
	Cdc.RegisterConcrete(sr25519.PubKeySr25519{},
		sr25519.PubKeyAminoName, nil)
	Cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{},
		secp256k1.PubKeyAminoName, nil)
}
