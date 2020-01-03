package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/exported"
)

var ModuleCdc = codec.New()

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgGrantAuthorization{}, "cosmos-sdk/GrantAuthorization", nil)
	cdc.RegisterConcrete(MsgRevokeAuthorization{}, "cosmos-sdk/RevokeAuthorization", nil)
	cdc.RegisterConcrete(MsgExecDelegated{}, "cosmos-sdk/ExecDelegated", nil)
	cdc.RegisterConcrete(SendAuthorization{}, "cosmos-sdk/SendAuthorization", nil)
	cdc.RegisterConcrete(AuthorizationGrant{}, "cosmos-sdk/AuthorizationGrant", nil)
	cdc.RegisterConcrete(GenericAuthorization{}, "cosmos-sdk/GenericAuthorization", nil)

	cdc.RegisterInterface((*exported.Authorization)(nil), nil)
}

func init() {
	RegisterCodec(ModuleCdc)
}
