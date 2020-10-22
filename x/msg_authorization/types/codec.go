package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// RegisterLegacyAminoCodec registers concrete types and interfaces on the given codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Authorization)(nil), nil)
	cdc.RegisterInterface((*sdk.MsgRequest)(nil), nil)
	cdc.RegisterConcrete(&MsgGrantAuthorization{}, "cosmos-sdk/MsgGrantAuthorization", nil)
	cdc.RegisterConcrete(&MsgRevokeAuthorization{}, "cosmos-sdk/MsgRevokeAuthorization", nil)
	cdc.RegisterConcrete(&MsgExecAuthorized{}, "cosmos-sdk/MsgExecAuthorized", nil)
	cdc.RegisterConcrete(SendAuthorization{}, "cosmos-sdk/SendAuthorization", nil)
	cdc.RegisterConcrete(GenericAuthorization{}, "cosmos-sdk/GenericAuthorization", nil)
	cdc.RegisterConcrete(&banktypes.MsgSend{}, "cosmos-sdk/MsgSend", nil)
}

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgGrantAuthorization{},
		&MsgRevokeAuthorization{},
		&MsgExecAuthorized{},
	)

	registry.RegisterInterface(
		"cosmos.msg_authorization.v1beta1.Authorization",
		(*Authorization)(nil),
		&SendAuthorization{},
		&GenericAuthorization{},
	)

}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/msg_authorization module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/msg_authorization and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
}
