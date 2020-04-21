package simulation

import (
	"bytes"
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding slashing type.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
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
			var pubKeyA, pubKeyB gogotypes.StringValue
			cdc.MustUnmarshalBinaryBare(kvA.Value, &pubKeyA)
			cdc.MustUnmarshalBinaryBare(kvB.Value, &pubKeyB)

			bechPKA := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pubKeyA.Value)
			bechPKB := sdk.MustGetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pubKeyB.Value)
			return fmt.Sprintf("PubKeyA: %s\nPubKeyB: %s", bechPKA, bechPKB)

		default:
			panic(fmt.Sprintf("invalid slashing key prefix %X", kvA.Key[:1]))
		}
	}
}
