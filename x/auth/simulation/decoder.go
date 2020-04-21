package simulation

import (
	"bytes"
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding auth type.
func NewDecodeStore(cdc types.Codec) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
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
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
