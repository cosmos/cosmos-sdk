package types

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&authz.MsgGrant{},
		&authz.MsgRevoke{},
		&authz.MsgExec{},
	)

	registry.RegisterInterface(
		"cosmos.authz.v1beta1.Authorization",
		(*authz.Authorization)(nil),
		&authz.GenericAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, authz.MsgServiceDesc())
}
