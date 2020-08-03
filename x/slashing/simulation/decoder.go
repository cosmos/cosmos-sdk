package simulation

import (
	"bytes"
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding slashing type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.ValidatorSigningInfoKeyPrefix):
			var infoA, infoB types.ValidatorSigningInfo
			cdc.MustUnmarshalBinaryBare(kvA.Value, &infoA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &infoB)
			return fmt.Sprintf("%v\n%v", infoA, infoB)

		case bytes.Equal(kvA.Key[:1], types.ValidatorMissedBlockBitArrayKeyPrefix):
			var missedA, missedB gogotypes.BoolValue
			cdc.MustUnmarshalBinaryBare(kvA.Value, &missedA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &missedB)
			return fmt.Sprintf("missedA: %v\nmissedB: %v", missedA.Value, missedB.Value)

		case bytes.Equal(kvA.Key[:1], types.AddrPubkeyRelationKeyPrefix):
			var pubKeyA, pubKeyB gogotypes.StringValue
			cdc.MustUnmarshalBinaryBare(kvA.Value, &pubKeyA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &pubKeyB)
			return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", pubKeyA.Value, pubKeyB.Value)

		default:
			panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
		}
	}
}
