package simulation

import (
	"bytes"
	"fmt"

	"github.com/pointnetwork/cosmos-point-sdk/codec"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/types/kv"
	"github.com/pointnetwork/cosmos-point-sdk/x/capability/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding capability type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.Equal(kvA.Key, types.KeyIndex):
			idxA := sdk.BigEndianToUint64(kvA.Value)
			idxB := sdk.BigEndianToUint64(kvB.Value)
			return fmt.Sprintf("Index A: %d\nIndex B: %d\n", idxA, idxB)

		case bytes.HasPrefix(kvA.Key, types.KeyPrefixIndexCapability):
			var capOwnersA, capOwnersB types.CapabilityOwners
			cdc.MustUnmarshal(kvA.Value, &capOwnersA)
			cdc.MustUnmarshal(kvB.Value, &capOwnersB)
			return fmt.Sprintf("CapabilityOwners A: %v\nCapabilityOwners B: %v\n", capOwnersA, capOwnersB)

		default:
			panic(fmt.Sprintf("invalid %s key prefix %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
		}
	}
}
