package ibc

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/tendermint"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
}

func RegisterCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
	tendermint.RegisterCodec(cdc)
	channel.RegisterCodec(cdc)
	commitment.RegisterCodec(cdc)
	merkle.RegisterCodec(cdc)

	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/MsgUpdateClient", nil)
	cdc.RegisterConcrete(MsgOpenConnection{}, "ibc/MsgOpenConnection", nil)
	cdc.RegisterConcrete(MsgOpenChannel{}, "ibc/MsgOpenChannel", nil)
}
