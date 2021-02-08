package simulation

import (
	"bytes"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

//TODO: check decode store
// NewDecodeStore returns a function closure that unmarshals the KVPair's values
// to the corresponding types.
func NewDecodeStore(cdc codec.Marshaler) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.SupplyKey):
			var coinsA sdk.Coin
			err := cdc.UnmarshalBinaryBare(kvA.Value, &coinsA)
			if err != nil {
				panic(err)
			}

			var coinsB sdk.Coin
			err = cdc.UnmarshalBinaryBare(kvB.Value, &coinsB)
			if err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", coinsA, coinsA)

		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
