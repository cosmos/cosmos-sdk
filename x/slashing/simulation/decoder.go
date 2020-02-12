package simulation

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	gogotypes "github.com/gogo/protobuf/types"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/x/slashing/internal/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding slashing type
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.ValidatorSigningInfoKey):
		var infoA, infoB types.ValidatorSigningInfo
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &infoA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &infoB)
		return fmt.Sprintf("%v\n%v", infoA, infoB)

	case bytes.Equal(kvA.Key[:1], types.ValidatorMissedBlockBitArrayKey):
		var missedA, missedB gogotypes.BoolValue
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &missedA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &missedB)
		return fmt.Sprintf("missedA: %v missedB: %v\n", missedA.Value, missedB.Value)

	case bytes.Equal(kvA.Key[:1], types.AddrPubkeyRelationKey):
		var pubKeyA, pubKeyB gogotypes.StringValue
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &pubKeyA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &pubKeyB)

		return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", pubKeyA.Value, pubKeyB.Value)

	default:
		panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
	}
}
