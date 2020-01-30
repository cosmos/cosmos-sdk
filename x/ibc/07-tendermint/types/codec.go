package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// SubModuleCdc defines the IBC tendermint client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the Tendermint types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/tendermint/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/tendermint/Evidence", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc tendermint client codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}
