package client

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var MsgCdc = codec.New()

func init() {
	RegisterCodec(MsgCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ConsensusState)(nil), nil)
	cdc.RegisterInterface((*Header)(nil), nil)

	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/MsgUpdateClient", nil)
}
