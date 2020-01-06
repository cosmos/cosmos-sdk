package tendermint

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var SubModuleCdc = codec.New()

// RegisterCodec registers the Tendermint types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/tendermint/ClientState", nil)
	cdc.RegisterConcrete(Committer{}, "ibc/client/tendermint/Committer", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/tendermint/Evidence", nil)
}

func init() {
	RegisterCodec(SubModuleCdc)
}
