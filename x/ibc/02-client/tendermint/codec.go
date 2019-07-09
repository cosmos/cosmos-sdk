package tendermint

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

func init() {
	RegisterCodec(client.MsgCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
}
