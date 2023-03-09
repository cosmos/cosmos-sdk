package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/circuit/types"
	"github.com/stretchr/testify/require"
)

func Test_msgServer_AuthorizeCircuitBreaker(t *testing.T) {

	ft := SetupFixture(t)

	srv := msgServer{
		Keeper: ft.Keeper,
	}

	// add a new super admin
	adminPerms := &types.Permissions{Level: types.Permissions_LEVEL_SUPER_ADMIN, LimitTypeUrls: []string{""}}
	msg := &types.MsgAuthorizeCircuitBreaker{Granter: addresses[0], Grantee: addresses[1], Permissions: adminPerms}
	_, err := srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.NoError(t, err)

	add1, err := ft.Keeper.addressCodec.StringToBytes(addresses[1])
	require.NoError(t, err)

	perms, err := ft.Keeper.GetPermissions(ft.Ctx, add1)
	require.NoError(t, err)

	require.Equal(t, adminPerms, perms, "admin perms are not the same")

	// add a super user
	allmsgs := &types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{""}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[0], Grantee: addresses[2], Permissions: allmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.NoError(t, err)

	add2, err := ft.Keeper.addressCodec.StringToBytes(addresses[2])
	require.NoError(t, err)

	perms, err = ft.Keeper.GetPermissions(ft.Ctx, add2)
	require.NoError(t, err)

	require.Equal(t, allmsgs, perms, "admin perms are not the same")

	// unauthorized user who does not have perms trying to authorize
	superPerms := &types.Permissions{Level: types.Permissions_LEVEL_SUPER_ADMIN, LimitTypeUrls: []string{}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[3], Grantee: addresses[2], Permissions: superPerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.Error(t, err, "user with no permission should fail in authorizing others")

	// user with permission level all_msgs tries to grant another user perms
	somePerms := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[2], Grantee: addresses[3], Permissions: somePerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.Error(t, err, "user[2] does not have permission to grant others permission")

	// user successfully grants another user perms to a specific permission

	somemsgs := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{"cosmos.bank.v1beta1.MsgSend"}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[0], Grantee: addresses[3], Permissions: somemsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.NoError(t, err)

	add3, err := ft.Keeper.addressCodec.StringToBytes(addresses[3])
	require.NoError(t, err)

	perms, err = ft.Keeper.GetPermissions(ft.Ctx, add3)
	require.NoError(t, err)

	require.Equal(t, somemsgs, perms, "admin perms are not the same")

	// admin tries grants another user permission ALL_MSGS with limited urls populated
	invalidmsgs := &types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{"cosmos.bank.v1beta1.MsgSend"}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[0], Grantee: addresses[4], Permissions: invalidmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.Ctx, msg)
	require.Error(t, err)
}

// func Test_msgServer_TripCircuitBreaker(t *testing.T) {

// }

// func Test_msgServer_ResetCircuitBreaker(t *testing.T) {

// }
