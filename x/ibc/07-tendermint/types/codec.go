package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
)

// SubModuleCdc defines the IBC tendermint client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the Tendermint types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/tendermint/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/tendermint/Evidence", nil)
	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/tendermint/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/tendermint/MsgUpdateClient", nil)
	cdc.RegisterConcrete(MsgSubmitClientMisbehaviour{}, "ibc/client/tendermint/MsgSubmitClientMisbehaviour", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc tendermint client codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}

// RegisterInterfaces registers the tendermint concrete evidence and client-related
// implementations and interfaces.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*evidenceexported.Evidence)(nil),
		&Evidence{},
	)
}
