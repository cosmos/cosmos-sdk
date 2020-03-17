package simulation

import (
	"bytes"
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/tendermint/tendermint/crypto"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding slashing type
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.ValidatorSigningInfoKey):
		var infoA, infoB types.ValidatorSigningInfo
		cdc.MustUnmarshalBinaryBare(kvA.Value, &infoA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorMissedBlockBitArrayKey):
		var missedA, missedB gogotypes.BoolValue
		cdc.MustUnmarshalBinaryBare(kvA.Value, &missedA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &missedB)
		return fmt.Sprintf("missedA: %v\nmissedB: %v", missedA.Value, missedB.Value)

	case bytes.Equal(kvA.Key[:1], types.AddrPubkeyRelationKey):
		var pubKeyA, pubKeyB crypto.PubKey
		cdc.MustUnmarshalBinaryBare(kvA.Value, &pubKeyA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &pubKeyB)
		bechPKA := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKeyA)
		bechPKB := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeAccPub, pubKeyB)
		return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPKA, bechPKB)

	default:
		panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
	}
}
