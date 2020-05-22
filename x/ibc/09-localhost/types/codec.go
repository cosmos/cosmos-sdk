package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

const (
	// SubModuleName for the localhost (loopback) client
	SubModuleName = "localhost"
)

// SubModuleCdc defines the IBC localhost client codec.
var SubModuleCdc *codec.Codec

// RegisterCodec registers the localhost types
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/localhost/ClientState", nil)
	cdc.RegisterConcrete(MsgCreateClient{}, "ibc/client/localhost/MsgCreateClient", nil)
	SetSubModuleCodec(cdc)
}

// SetSubModuleCodec sets the ibc localhost client codec
func SetSubModuleCodec(cdc *codec.Codec) {
	SubModuleCdc = cdc
}

// RegisterInterfaces registers the solo machine concrete evidence and client-related
// implementations and interfaces.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateClient{},
	)
	registry.RegisterImplementations(
		(*clientexported.ClientState)(nil),
		&ClientState{},
	)
}
