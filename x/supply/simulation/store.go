package simulation

import (
	"bytes"
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/supply"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding supply type
func DecodeStore(cdcA, cdcB *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], supply.SupplyKey):
		var supplyA, supplyB supply.Supply
		cdcA.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
		cdcB.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
		return fmt.Sprintf("%v\n%v", supplyB, supplyB)
	default:
		panic(fmt.Sprintf("invalid supply key %X", kvA.Key))
	}
}
