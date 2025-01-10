package simulation

import (
	"bytes"
	"fmt"

	"cosmossdk.io/core/codec"
	"cosmossdk.io/x/authz"
	"cosmossdk.io/x/authz/keeper"

	"github.com/cosmos/cosmos-sdk/types/kv"
)

// NewDecodeStore returns a decoder function closure that umarshals the KVPair's
// Value to the corresponding authz type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key[:1], keeper.GrantKey):
			var grantA, grantB authz.Grant
			if err := cdc.Unmarshal(kvA.Value, &grantA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &grantB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", grantA, grantB)
		case bytes.Equal(kvA.Key[:1], keeper.GrantQueuePrefix):
			var grantA, grantB authz.GrantQueueItem
			if err := cdc.Unmarshal(kvA.Value, &grantA); err != nil {
				panic(err)
			}
			if err := cdc.Unmarshal(kvB.Value, &grantB); err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", grantA, grantB)
		default:
			panic(fmt.Sprintf("invalid authz key %X", kvA.Key))
		}
	}
}
