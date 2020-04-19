package simulation

import (
	"bytes"
	"fmt"

	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// DecodeStore unmarshals the KVPair's values to the corresponding types.
func DecodeStore(cdc *codec.Codec, kvA, kvB tmkv.Pair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], types.SupplyKey):
		var supplyA, supplyB types.Supply
		cdc.MustUnmarshalBinaryBare(kvA.Value, &supplyA)
		cdc.MustUnmarshalBinaryBare(kvB.Value, &supplyB)
		return fmt.Sprintf("%v\n%v", supplyB, supplyB)

	default:
		panic(fmt.Sprintf("unknown %s key %X (%s)", types.ModuleName, kvA.Key, kvA.Key))
	}
}
