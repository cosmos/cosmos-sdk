package types

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.MsgRequest)(nil),
		&authz.MsgGrant{},
		&authz.MsgRevoke{},
		&authz.MsgExec{},
	)

	registry.RegisterInterface(
		"cosmos.authz.v1beta1.Authorization",
		(*authz.Authorization)(nil),
		&bank.SendAuthorization{},
		&authz.GenericAuthorization{},
		&staking.StakeAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, authz.MsgServiceDesc())
}
