package multisig

import (
	"github.com/cosmos/cosmos-sdk/crypto/types"
	amino "github.com/tendermint/go-amino"

	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/crypto/sr25519"
)

// TODO: Figure out API for others to either add their own pubkey types, or
// to make verify / marshal accept a cdc.
const (
	PubKeyAminoRoute = "tendermint/PubKeyMultisigThreshold"
)

var cdc = amino.NewCodec()

func init() {
	cdc.RegisterInterface((*crypto.PubKey)(nil), nil)
	cdc.RegisterConcrete(PubKeyMultisigThreshold{},
		PubKeyAminoRoute, nil)
	cdc.RegisterConcrete(ed25519.PubKeyEd25519{},
		ed25519.PubKeyAminoName, nil)
	cdc.RegisterConcrete(sr25519.PubKeySr25519{},
		sr25519.PubKeyAminoName, nil)
	cdc.RegisterConcrete(secp256k1.PubKeySecp256k1{},
		secp256k1.PubKeyAminoName, nil)
}

// AminoMultisignature is used to represent the signature object used in the multisigs.
// Sigs is a list of signatures, sorted by corresponding index.
type AminoMultisignature struct {
	BitArray *types.CompactBitArray
	Sigs     [][]byte
}
