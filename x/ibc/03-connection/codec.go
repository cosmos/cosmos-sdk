package connection

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

var MsgCdc = codec.New()

func init() {
	commitment.RegisterCodec(MsgCdc)
	merkle.RegisterCodec(MsgCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgOpenInit{}, "cosmos-sdk/ibc/connection/MsgOpenInit", nil)
	cdc.RegisterConcrete(MsgOpenTry{}, "cosmos-sdk/ibc/connection/MsgOpenTry", nil)
	cdc.RegisterConcrete(MsgOpenAck{}, "cosmos-sdk/ibc/connection/MsgOpenAck", nil)
	cdc.RegisterConcrete(MsgOpenConfirm{}, "cosmos-sdk/ibc/connection/MsgOpenConfirm", nil)
}
