package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// SubModuleCdc defines the IBC client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the IBC client interfaces and types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.ClientState)(nil), nil)
	cdc.RegisterInterface((*exported.MsgCreateClient)(nil), nil)
	cdc.RegisterInterface((*exported.MsgUpdateClient)(nil), nil)
	cdc.RegisterInterface((*exported.ConsensusState)(nil), nil)
	cdc.RegisterInterface((*exported.Header)(nil), nil)
	cdc.RegisterInterface((*exported.Misbehaviour)(nil), nil)

	SetSubModuleCodec(cdc)
}

func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
