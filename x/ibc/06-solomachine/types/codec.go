package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// SubModuleCdc defines the IBC solo machine client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the Solo Machine types.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/solomachine/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/solomachine/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/solomachine/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/solomachine/Evidence", nil)
	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/solomachine/MsgCreateClient", nil)
	cdc.RegisterConcrete(MsgUpdateClient{}, "ibc/client/solomachine/MsgUpdateClient", nil)
	cdc.RegisterConcrete(MsgSubmitClientMisbehaviour{}, "ibc/client/solomachine/MsgSubmitClientMisbehaviour", nil)

	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc solo machine client codec.
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}

// RegisterInterfaces registers the solo machine concrete evidence and client-related
// implementations and interfaces.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateClient{},
		&MsgUpdateClient{},
		&MsgSubmitClientMisbehaviour{},
	)
	registry.RegisterImplementations(
		(*clientexported.ClientState)(nil),
		&ClientState{},
	)
	registry.RegisterImplementations(
		(*clientexported.ConsensusState)(nil),
		&ConsensusState{},
	)
	registry.RegisterImplementations(
		(*clientexported.Header)(nil),
		&Header{},
	)
	registry.RegisterImplementations(
		(*clientexported.Misbehaviour)(nil),
		&Evidence{},
	)
	registry.RegisterImplementations(
		(*evidenceexported.Evidence)(nil),
		&Evidence{},
	)
}
