package simulation

import (
	"bytes"
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding auth type
func DecodeStore(cdcI interface{}, kvA, kvB tmkv.Pair) string {
	cdc, ok := cdcI.(types.Codec)
	if !ok {
		panic(fmt.Sprintf("invalid codec: %T", cdcI))
	}

	switch {
	case bytes.Equal(kvA.Key[:1], types.AddressStoreKeyPrefix):
		accA, err := cdc.UnmarshalAccount(kvA.Value)
		if err != nil {
			panic(err)
		}

		accB, err := cdc.UnmarshalAccount(kvB.Value)
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("%v\n%v", accA, accB)

	case bytes.Equal(kvA.Key, types.GlobalAccountNumberKey):
		var globalAccNumberA, globalAccNumberB gogotypes.UInt64Value
		cdc.MustUnmarshalBinaryBare(kvA.Value, &globalAccNumberA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &globalAccNumberB)
		return fmt.Sprintf("GlobalAccNumberA: %d\nGlobalAccNumberB: %d", globalAccNumberA, globalAccNumberB)

	default:
		panic(fmt.Sprintf("invalid account key %X", kvA.Key))
	}
}
