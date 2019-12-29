package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
)

type Keeper interface {
	//DispatchActions executes the provided messages via authorization grants from the message signer to the grantee
	DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) sdk.Result

	// Grants the provided authorization to the grantee on the granter's account with the provided expiration time
	// If there is an existing authorization grant for the same sdk.Msg type, this grant overwrites that.
	Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, authorization types.Authorization, expiration time.Time)

	//Revokes any authorization for the provided message type granted to the grantee by the granter.
	Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg)

	//Returns any Authorization (or nil), with the expiration time,
	// granted to the grantee by the granter for the provided msg type.
	GetAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) (cap types.Authorization, expiration time.Time)
}
