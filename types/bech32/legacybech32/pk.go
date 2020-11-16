// Deprecated: The module provides legacy bech32 functions which will be removed in a future
// release.
package legacybech32

// nolint

// TODO: remove Bech32 prefix, it's already in package

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// Deprecated: Bech32PubKeyType defines a string type alias for a Bech32 public key type.
type Bech32PubKeyType string

// Bech32 conversion constants
const (
	Bech32PubKeyTypeAccPub  Bech32PubKeyType = "accpub"
	Bech32PubKeyTypeValPub  Bech32PubKeyType = "valpub"
	Bech32PubKeyTypeConsPub Bech32PubKeyType = "conspub"
)

// Deprecated: Bech32ifyPubKey returns a Bech32 encoded string containing the appropriate
// prefix based on the key type provided for a given PublicKey.
func Bech32ifyPubKey(pkt Bech32PubKeyType, pubkey cryptotypes.PubKey) (string, error) {
	bech32Prefix := getPrefix(pkt)
	return bech32.ConvertAndEncode(bech32Prefix, legacy.Cdc.MustMarshalBinaryBare(pubkey))
}

// Deprecated: MustBech32ifyPubKey calls Bech32ifyPubKey and panics on error.
func MustBech32ifyPubKey(pkt Bech32PubKeyType, pubkey cryptotypes.PubKey) string {
	res, err := Bech32ifyPubKey(pkt, pubkey)
	if err != nil {
		panic(err)
	}

	return res
}

func getPrefix(pkt Bech32PubKeyType) string {
	cfg := sdk.GetConfig()
	switch pkt {
	case Bech32PubKeyTypeAccPub:
		return cfg.GetBech32AccountPubPrefix()

	case Bech32PubKeyTypeValPub:
		return cfg.GetBech32ValidatorPubPrefix()
	case Bech32PubKeyTypeConsPub:
		return cfg.GetBech32ConsensusPubPrefix()
	}

	return ""
}

// Deprecated: GetPubKeyFromBech32 returns a PublicKey from a bech32-encoded PublicKey with
// a given key type.
func GetPubKeyFromBech32(pkt Bech32PubKeyType, pubkeyStr string) (cryptotypes.PubKey, error) {
	bech32Prefix := getPrefix(pkt)

	bz, err := sdk.GetFromBech32(pubkeyStr, bech32Prefix)
	if err != nil {
		return nil, err
	}

	return cryptocodec.PubKeyFromBytes(bz)
}

// Deprecated: MustGetPubKeyFromBech32 calls GetPubKeyFromBech32 except it panics on error.
func MustGetPubKeyFromBech32(pkt Bech32PubKeyType, pubkeyStr string) cryptotypes.PubKey {
	res, err := GetPubKeyFromBech32(pkt, pubkeyStr)
	if err != nil {
		panic(err)
	}

	return res
}
