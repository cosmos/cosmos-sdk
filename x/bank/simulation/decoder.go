package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type SupplyUnmarshaller interface {
	UnmarshalSupply([]byte) (exported.SupplyI, error)
}

// NewDecodeStore returns a function closure that unmarshals the KVPair's values
// to the corresponding types.
func NewDecodeStore(cdc SupplyUnmarshaller) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], types.SupplyKey):
			supplyA, err := cdc.UnmarshalSupply(kvA.Value)
			if err != nil {
				panic(err)
			}

			supplyB, err := cdc.UnmarshalSupply(kvB.Value)
			if err != nil {
				panic(err)
			}

			return fmt.Sprintf("%v\n%v", supplyA, supplyB)

		default:
			panic(fmt.Sprintf("unexpected %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
