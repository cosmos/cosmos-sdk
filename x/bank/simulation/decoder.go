package simulation

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewDecodeStore returns a function closure that unmarshals the KVPair's values
// to the corresponding types.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB tmkv.Pair) string {
	return func(kvA, kvB tmkv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.SupplyKey):
			var supplyA types.Supply
			err := cdc.UnmarshalBinaryBare(kvA.Value, &supplyA)
			if err != nil {
				panic(err)
			}

			var supplyB types.Supply
			err = cdc.UnmarshalBinaryBare(kvB.Value, &supplyB)
			if err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", supplyA, supplyB)

		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
