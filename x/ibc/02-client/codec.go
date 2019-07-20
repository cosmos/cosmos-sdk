package client

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

var MsgCdc = codec.New()

func init() {
	codec.RegisterCrypto(MsgCdc)
	commitment.RegisterCodec(MsgCdc)
	merkle.RegisterCodec(MsgCdc)
	RegisterCodec(MsgCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ConsensusState)(nil), nil)
	cdc.RegisterInterface((*Header)(nil), nil)

	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/MsgUpdateClient", nil)
}
