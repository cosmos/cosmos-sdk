package exported

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper interface {
	// DispatchActions executes the provided messages via authorization grants from the message signer to the grantee
	DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.ServiceMsg) sdk.Result

	// Grants the provided authorization to the grantee on the granter's account with the provided expiration time
	// If there is an existing authorization grant for the same sdk.Msg type, this grant overwrites that.
	Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, authorization Authorization, expiration time.Time) error

	// Revokes any authorization for the provided message type granted to the grantee by the granter.
	Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string)

	// Returns any Authorization (or nil), with the expiration time,
	// granted to the grantee by the granter for the provided msg type.
	// If the Authorization is expired already, it will revoke the authorization and return nil
	GetOrRevokeAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (cap Authorization, expiration time.Time)
}
