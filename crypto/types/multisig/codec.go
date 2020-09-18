package multisig

import (
	amino "github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/sr25519"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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
	Cdc.RegisterConcrete(ed25519.PubKey{},
		ed25519.PubKeyName, nil)
	Cdc.RegisterConcrete(sr25519.PubKey{},
		sr25519.PubKeyName, nil)
	Cdc.RegisterConcrete(&secp256k1.PubKey{},
		secp256k1.PubKeyName, nil)
}
