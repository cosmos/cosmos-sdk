package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var ProtoCodec *codec.ProtoCodec

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	ProtoCodec = codec.NewProtoCodec(registry)
	registry.RegisterImplementations((*sdk.MsgRequest)(nil),
		&MsgGrantAuthorization{},
		&MsgRevokeAuthorization{},
		&MsgExecAuthorized{},
	)

	registry.RegisterInterface(
		"cosmos.authz.v1beta1.Authorization",
		(*Authorization)(nil),
		&SendAuthorization{},
		&GenericAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
