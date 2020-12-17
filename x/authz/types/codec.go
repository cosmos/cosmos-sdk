package types

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// Registry is interface
var Registry types.InterfaceRegistry

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	Registry = registry
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
