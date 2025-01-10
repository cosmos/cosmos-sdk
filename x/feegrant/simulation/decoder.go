package simulation

import (
	"bytes"
	"fmt"

	"cosmossdk.io/core/codec"
	"cosmossdk.io/x/feegrant"

	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding feegrant type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], feegrant.FeeAllowanceKeyPrefix):
			var grantA, grantB feegrant.Grant
			if err := cdc.Unmarshal(kvA.Value, &grantA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &grantB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", grantA, grantB)
		default:
			panic(fmt.Sprintf("invalid feegrant key %X", kvA.Key))
		}
	}
}
