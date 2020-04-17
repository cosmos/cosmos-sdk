package localhost

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// SubModuleCdc defines the IBC localhost client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the localhost types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/localhost/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/localhost/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/localhost/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/localhost/Evidence", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc localhost client codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
