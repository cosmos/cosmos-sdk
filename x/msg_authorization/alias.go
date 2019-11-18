package msg_authorization

import (
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	QuerierRoute = types.QuerierRoute
)

var (
	NewKeeper = keeper.NewKeeper
)

type (
	Keeper                 = keeper.Keeper
	Capability             = types.Capability
	MsgGrantAuthorization  = types.MsgGrantAuthorization
	MsgRevokeAuthorization = types.MsgRevokeAuthorization
	MsgExecDelegated       = types.MsgExecDelegated
)
