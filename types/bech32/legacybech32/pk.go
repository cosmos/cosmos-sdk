// Deprecated: The module provides legacy bech32 functions which will be removed in a future
// release.
package legacybech32

// nolint

// TODO: remove Bech32 prefix, it's already in package

import (
	"github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
func Bech32ifyPubKey(pkt Bech32PubKeyType, pubkey crypto.PubKey) (string, error) {
	// This piece of code is to keep backwards-compatibility.
	// For ed25519 keys, our own ed25519 is registered in Amino under a
	// different name than TM's ed25519. But since users are already using
	// TM's ed25519 bech32 encoding, we explicitly say to bech32-encode our own
	// ed25519 the same way as TM's ed25519.
	// TODO: Remove Bech32ifyPubKey and all usages (cosmos/cosmos-sdk/issues/#7357)
	pkToMarshal := pubkey
	if ed25519Pk, ok := pubkey.(*ed25519.PubKey); ok {
		pkToMarshal = ed25519Pk.AsTmPubKey()
	}

	bech32Prefix := getPrefix(pkt)
	return bech32.ConvertAndEncode(bech32Prefix, legacy.Cdc.MustMarshalBinaryBare(pkToMarshal))
}

// Deprecated: MustBech32ifyPubKey calls Bech32ifyPubKey and panics on error.
func MustBech32ifyPubKey(pkt Bech32PubKeyType, pubkey crypto.PubKey) string {
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
func GetPubKeyFromBech32(pkt Bech32PubKeyType, pubkeyStr string) (crypto.PubKey, error) {
	bech32Prefix := getPrefix(pkt)
	bz, err := sdk.GetFromBech32(pubkeyStr, bech32Prefix)
	if err != nil {
		return nil, err
	}

	aminoPk, err := cryptocodec.PubKeyFromBytes(bz)
	if err != nil {
		return nil, err
	}

	var protoPk crypto.PubKey
	switch aminoPk.(type) {

	// We are bech32ifying some secp256k1 keys in tests.
	case *secp256k1.PubKey:
		protoPk = aminoPk
	case *ed25519.PubKey:
		protoPk = aminoPk

	// Real-life case.
	case tmed25519.PubKey:
		protoPk = &ed25519.PubKey{
			Key: aminoPk.Bytes(),
		}

	default:
		// We only allow ed25519 pubkeys to be bech32-ed right now.
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "bech32 pubkey does not support %T", aminoPk)
	}

	return protoPk, nil
}

// Deprecated: MustGetPubKeyFromBech32 calls GetPubKeyFromBech32 except it panics on error.
func MustGetPubKeyFromBech32(pkt Bech32PubKeyType, pubkeyStr string) crypto.PubKey {
	res, err := GetPubKeyFromBech32(pkt, pubkeyStr)
	if err != nil {
		panic(err)
	}

	return res
}
