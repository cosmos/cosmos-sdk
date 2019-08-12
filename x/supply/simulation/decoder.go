package simulation

import (
	"bytes"
	"fmt"

	cmn "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/internal/types"
)

// DecodeStore unmarshals the KVPair's Value to the corresponding supply type
func DecodeStore(cdc *codec.Codec, kvA, kvB cmn.KVPair) string {
	switch {
	case bytes.Equal(kvA.Key[:1], keeper.SupplyKey):
		var supplyA, supplyB types.Supply
		cdc.MustUnmarshalBinaryLengthPrefixed(kvA.Value, &supplyA)
		cdc.MustUnmarshalBinaryLengthPrefixed(kvB.Value, &supplyB)
		return fmt.Sprintf("%v\n%v", supplyB, supplyB)
	default:
		panic(fmt.Sprintf("invalid supply key %X", kvA.Key))
	}
}
