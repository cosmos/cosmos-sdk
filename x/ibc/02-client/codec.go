package client

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var MsgCdc *codec.Codec

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ConsensusState)(nil), nil)
	cdc.RegisterInterface((*Header)(nil), nil)

	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/MsgUpdateClient", nil)
}

func SetMsgCodec(cdc *codec.Codec) {
	// TODO
	/*
		if MsgCdc != nil && MsgCdc != cdc {
			panic("MsgCdc set more than once")
		}
	*/
	MsgCdc = cdc
}
