package tendermint

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)

}
