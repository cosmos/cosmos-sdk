package bech32pubkey

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// Bech32PubKeyType defines a string type alias for a Bech32 public key type.
type Bech32PubKeyType string

// Bech32 conversion constants
const (
	Bech32PubKeyTypeAccPub  Bech32PubKeyType = "accpub"
	Bech32PubKeyTypeValPub  Bech32PubKeyType = "valpub"
	Bech32PubKeyTypeConsPub Bech32PubKeyType = "conspub"
)

// Bech32ifyPubKey returns a Bech32 encoded string containing the appropriate
// prefix based on the key type provided for a given PublicKey.
// TODO: Remove Bech32ifyPubKey and all usages (cosmos/cosmos-sdk/issues/#7357)
func Bech32ifyPubKey(pkt Bech32PubKeyType, pubkey cryptotypes.PubKey) (string, error) {
	var bech32Prefix string

	switch pkt {
	case Bech32PubKeyTypeAccPub:
		bech32Prefix = sdk.GetConfig().GetBech32AccountPubPrefix()

	case Bech32PubKeyTypeValPub:
		bech32Prefix = sdk.GetConfig().GetBech32ValidatorPubPrefix()

	case Bech32PubKeyTypeConsPub:
		bech32Prefix = sdk.GetConfig().GetBech32ConsensusPubPrefix()

	}

	return bech32.ConvertAndEncode(bech32Prefix, legacy.Cdc.Amino.MustMarshalBinaryBare(pubkey))
}

// MustBech32ifyPubKey calls Bech32ifyPubKey except it panics on error.
func MustBech32ifyPubKey(pkt Bech32PubKeyType, pubkey cryptotypes.PubKey) string {
	res, err := Bech32ifyPubKey(pkt, pubkey)
	if err != nil {
		panic(err)
	}

	return res
}

// GetPubKeyFromBech32 returns a PublicKey from a bech32-encoded PublicKey with
// a given key type.
func GetPubKeyFromBech32(pkt Bech32PubKeyType, pubkeyStr string) (cryptotypes.PubKey, error) {
	var bech32Prefix string

	switch pkt {
	case Bech32PubKeyTypeAccPub:
		bech32Prefix = sdk.GetConfig().GetBech32AccountPubPrefix()

	case Bech32PubKeyTypeValPub:
		bech32Prefix = sdk.GetConfig().GetBech32ValidatorPubPrefix()

	case Bech32PubKeyTypeConsPub:
		bech32Prefix = sdk.GetConfig().GetBech32ConsensusPubPrefix()

	}

	bz, err := sdk.GetFromBech32(pubkeyStr, bech32Prefix)
	if err != nil {
		return nil, err
	}

	return legacy.PubKeyFromBytes(bz)
}

// MustGetPubKeyFromBech32 calls GetPubKeyFromBech32 except it panics on error.
func MustGetPubKeyFromBech32(pkt Bech32PubKeyType, pubkeyStr string) cryptotypes.PubKey {
	res, err := GetPubKeyFromBech32(pkt, pubkeyStr)
	if err != nil {
		panic(err)
	}

	return res
}
