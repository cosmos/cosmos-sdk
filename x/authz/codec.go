package authz

import (
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgGrant{},
		&MsgRevoke{},
		&MsgExec{},
	)

	// since bank.SendAuthorization and staking.StakeAuthorization both implement Authorization
	// and authz depends on x/bank and x/staking in other places, these registrations are placed here
	// to prevent a cyclic dependency.
	// see: https://github.com/cosmos/cosmos-sdk/pull/16509
	registry.RegisterInterface(
		"cosmos.authz.v1beta1.Authorization",
		(*Authorization)(nil),
		&GenericAuthorization{},
		&bank.SendAuthorization{},
		&staking.StakeAuthorization{},
	)
	msgservice.RegisterMsgServiceDesc(registry, MsgServiceDesc())
}
