package simulation

import (
	tmkv "github.com/tendermint/tendermint/libs/kv"

	"github.com/cosmos/cosmos-sdk/codec"
	channelsim "github.com/cosmos/cosmos-sdk/x/04-channel/simulation"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding ibc type.
func NewDecodeStore(cdc codec.Marshaler, aminoCdc *codec.Codec) func(kvA, kvB tmkv.Pair) string {
	//TODO: switch
	return channelsim.NewDecodeStore(cdc)
}
