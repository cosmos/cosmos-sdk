package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
)

const (
	// SubModuleName for the localhost (loopback) client
	SubModuleName = "localhost"
)

// SubModuleCdc defines the IBC localhost client codec.
var SubModuleCdc *codec.Codec

func init() {
	SubModuleCdc = codec.New()
	cryptocodec.RegisterCrypto(SubModuleCdc)
	RegisterCodec(SubModuleCdc)
}

// RegisterCodec registers the localhost types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/localhost/ClientState", nil)
	cdc.RegisterConcrete(&MsgCreateClient{}, "ibc/client/localhost/MsgCreateClient", nil)
}
