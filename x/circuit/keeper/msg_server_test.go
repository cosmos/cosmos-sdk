package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"
)

const msgSend = "cosmos.bank.v1beta1.MsgSend"

func Test_AuthorizeCircuitBreaker(t *testing.T) {
	ft := initFixture(t)

	srv := keeper.NewMsgServerImpl(ft.keeper)
	authority, err := ft.ac.BytesToString(ft.mockAddr)
	require.NoError(t, err)

	// add a new super admin
	adminPerms := types.Permissions{Level: types.Permissions_LEVEL_SUPER_ADMIN, LimitTypeUrls: []string{""}}
	msg := &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[1], Permissions: &adminPerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	add1, err := ft.ac.StringToBytes(addresses[1])
	require.NoError(t, err)

	perms, err := ft.keeper.Permissions.Get(ft.ctx, add1)
	require.NoError(t, err)

	require.Equal(t, adminPerms, perms, "admin perms are not the same")

	// add a super user
	allmsgs := types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{""}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[2], Permissions: &allmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	add2, err := ft.ac.StringToBytes(addresses[2])
	require.NoError(t, err)

	perms, err = ft.keeper.Permissions.Get(ft.ctx, add2)
	require.NoError(t, err)

	require.Equal(t, allmsgs, perms, "admin perms are not the same")

	// unauthorized user who does not have perms trying to authorize
	superPerms := &types.Permissions{Level: types.Permissions_LEVEL_SUPER_ADMIN, LimitTypeUrls: []string{}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[3], Grantee: addresses[2], Permissions: superPerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.Error(t, err, "user with no permission should fail in authorizing others")

	// user with permission level all_msgs tries to grant another user perms
	somePerms := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[2], Grantee: addresses[3], Permissions: somePerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.Error(t, err, "user[2] does not have permission to grant others permission")

	// user successfully grants another user perms to a specific permission

	somemsgs := types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{msgSend}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[3], Permissions: &somemsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	add3, err := ft.ac.StringToBytes(addresses[3])
	require.NoError(t, err)

	perms, err = ft.keeper.Permissions.Get(ft.ctx, add3)
	require.NoError(t, err)

	require.Equal(t, somemsgs, perms, "admin perms are not the same")

	add4, err := ft.ac.StringToBytes(addresses[4])
	require.NoError(t, err)

	perms, err = ft.keeper.Permissions.Get(ft.ctx, add4)
	require.ErrorIs(t, err, collections.ErrNotFound, "user should have no perms by default")

	require.Equal(t, types.Permissions{Level: types.Permissions_LEVEL_NONE_UNSPECIFIED, LimitTypeUrls: nil}, perms, "user should have no perms by default")

	// admin tries grants another user permission ALL_MSGS with limited urls populated
	invalidmsgs := types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{msgSend}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[4], Permissions: &invalidmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)
}

func Test_TripCircuitBreaker(t *testing.T) {
	ft := initFixture(t)

	srv := keeper.NewMsgServerImpl(ft.keeper)
	url := msgSend

	authority, err := ft.ac.BytesToString(ft.mockAddr)
	require.NoError(t, err)

	// admin trips circuit breaker
	admintrip := &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	allowed, err := ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	// user with all messages trips circuit breaker
	// add a super user
	allmsgs := &types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{""}}
	msg := &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[1], Permissions: allmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	// try to trip the circuit breaker
	url2 := "cosmos.staking.v1beta1.MsgDelegate"
	superTrip := &types.MsgTripCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, superTrip)
	require.NoError(t, err)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url2)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	// user with no permission attempts to trips circuit breaker
	unknownTrip := &types.MsgTripCircuitBreaker{Authority: addresses[4], MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, unknownTrip)
	require.Error(t, err)

	// user has permission to trip circuit breaker for two messages but only has permission for one
	url, url2 = "cosmos.staking.v1beta1.MsgCreateValidator", "cosmos.staking.v1beta1.MsgEditValidator"
	somemsgs := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{url}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[2], Permissions: somemsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	// try to trip two messages but user only has permission for one
	someTrip := &types.MsgTripCircuitBreaker{Authority: addresses[2], MsgTypeUrls: []string{url, url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, someTrip)
	require.ErrorContains(t, err, "MsgEditValidator")

	// user tries to trip an already tripped circuit breaker
	alreadyTripped := "cosmos.bank.v1beta1.MsgSend"
	twoTrip := &types.MsgTripCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{alreadyTripped}}
	_, err = srv.TripCircuitBreaker(ft.ctx, twoTrip)
	require.Error(t, err)
}

func Test_ResetCircuitBreaker(t *testing.T) {
	ft := initFixture(t)
	authority, err := ft.ac.BytesToString(ft.mockAddr)
	require.NoError(t, err)

	srv := keeper.NewMsgServerImpl(ft.keeper)

	// admin resets circuit breaker
	url := "cosmos.bank.v1beta1.MsgSend"
	// admin trips circuit breaker
	admintrip := &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	allowed, err := ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	adminReset := &types.MsgResetCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, adminReset)
	require.NoError(t, err)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.True(t, allowed, "circuit breaker should be reset")

	// user has no  permission to reset circuit breaker
	// admin trips circuit breaker
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	unknownUserReset := &types.MsgResetCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{url}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, unknownUserReset)
	require.Error(t, err)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be reset")

	// user with all messages resets circuit breaker
	allmsgs := &types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{""}}
	msg := &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[1], Permissions: allmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	//  trip the circuit breaker
	url2 := "cosmos.staking.v1beta1.MsgDelegate"
	admintrip = &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	// user with all messages resets circuit breaker
	allMsgsReset := &types.MsgResetCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{url}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, allMsgsReset)
	require.NoError(t, err)

	// user tries to reset an message they dont have permission to reset

	url = "cosmos.staking.v1beta1.MsgCreateValidator"
	// give restricted perms to a user
	someMsgs := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{url2}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[2], Permissions: someMsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	admintrip = &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	// user with all messages resets circuit breaker
	someMsgsReset := &types.MsgResetCircuitBreaker{Authority: addresses[2], MsgTypeUrls: []string{url}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, someMsgsReset)
	require.NoError(t, err)

	// user tries to reset an already reset circuit breaker
	admintrip = &types.MsgTripCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.Error(t, err)
}
