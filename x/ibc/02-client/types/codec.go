package types

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	cryptocodec "github.com/KiraCore/cosmos-sdk/crypto/codec"
	"github.com/KiraCore/cosmos-sdk/x/ibc/02-client/exported"
)

// SubModuleCdc defines the IBC client codec.
var SubModuleCdc *codec.Codec

func init() {
	SubModuleCdc = codec.New()
	cryptocodec.RegisterCrypto(SubModuleCdc)
	RegisterCodec(SubModuleCdc)
}

// RegisterCodec registers the IBC client interfaces and types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.ClientState)(nil), nil)
	cdc.RegisterInterface((*exported.MsgCreateClient)(nil), nil)
	cdc.RegisterInterface((*exported.MsgUpdateClient)(nil), nil)
	cdc.RegisterInterface((*exported.ConsensusState)(nil), nil)
	cdc.RegisterInterface((*exported.Header)(nil), nil)
	cdc.RegisterInterface((*exported.Misbehaviour)(nil), nil)
}
