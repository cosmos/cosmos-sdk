package authz

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	types "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterLegacyAminoCodec registers the necessary x/authz interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgGrant{}, "cosmos-sdk/MsgGrant")
	legacy.RegisterAminoMsg(cdc, &MsgRevoke{}, "cosmos-sdk/MsgRevoke")
	legacy.RegisterAminoMsg(cdc, &MsgExec{}, "cosmos-sdk/MsgExec")

	cdc.RegisterInterface((*Authorization)(nil), nil)
	cdc.RegisterConcrete(&GenericAuthorization{}, "cosmos-sdk/GenericAuthorization", nil)
}

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
