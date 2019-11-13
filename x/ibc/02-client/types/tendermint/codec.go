package tendermint

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var SubModuleCdc = codec.New()

// RegisterCodec registers the Tendermint types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(Misbehaviour{}, "ibc/client/tendermint/Misbehaviour", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/tendermint/Evidence", nil)
}

func init() {
	RegisterCodec(SubModuleCdc)
}
