package simulation

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding group type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch {
		case bytes.HasPrefix(kvA.Key, quarantine.OptInPrefix):
			return fmt.Sprintf("%v\n%v", kvA.Value, kvB.Value)

		case bytes.HasPrefix(kvA.Key, quarantine.AutoResponsePrefix):
			respA := quarantine.ToAutoResponse(kvA.Value)
			respB := quarantine.ToAutoResponse(kvB.Value)
			return fmt.Sprintf("%s\n%s", respA.String(), respB.String())

		case bytes.HasPrefix(kvA.Key, quarantine.RecordPrefix):
			var qrA, qrB quarantine.QuarantineRecord
			cdc.MustUnmarshal(kvA.Value, &qrA)
			cdc.MustUnmarshal(kvB.Value, &qrB)
			return fmt.Sprintf("%v\n%v", qrA, qrB)

		case bytes.HasPrefix(kvA.Key, quarantine.RecordIndexPrefix):
			var riA, riB quarantine.QuarantineRecordSuffixIndex
			cdc.MustUnmarshal(kvA.Value, &riA)
			cdc.MustUnmarshal(kvB.Value, &riB)
			return fmt.Sprintf("%v\n%v", riA, riB)

		default:
			panic(fmt.Sprintf("invalid quarantine key %X", kvA.Key))
		}
	}
}
