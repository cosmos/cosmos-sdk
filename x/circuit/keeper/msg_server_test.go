package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const msgSend = "cosmos.bank.v1beta1.MsgSend"

func TestAuthorizeCircuitBreaker(t *testing.T) {
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

	require.Equal(t, adminPerms, perms, "admin perms are the same")

	// add a super user
	allmsgs := types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{""}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[2], Permissions: &allmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewEvent(
			"authorize_circuit_breaker",
			sdk.NewAttribute("granter", authority),
			sdk.NewAttribute("grantee", addresses[2]),
			sdk.NewAttribute("permission", allmsgs.String()),
		),
		lastEvent(ft.ctx),
	)

	add2, err := ft.ac.StringToBytes(addresses[2])
	require.NoError(t, err)

	perms, err = ft.keeper.Permissions.Get(ft.ctx, add2)
	require.NoError(t, err)

	require.Equal(t, allmsgs, perms)

	// unauthorized user who does not have perms trying to authorize
	superPerms := &types.Permissions{Level: types.Permissions_LEVEL_SUPER_ADMIN, LimitTypeUrls: []string{}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[3], Grantee: addresses[2], Permissions: superPerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.Error(t, err, "user with no permission fails in authorizing others")

	// user with permission level all_msgs tries to grant another user perms
	somePerms := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: addresses[2], Grantee: addresses[3], Permissions: somePerms}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.Error(t, err, "super user[2] does not have permission to grant others permission")

	// admin successfully grants another user perms to a specific permission
	somemsgs := types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{msgSend}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[3], Permissions: &somemsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewEvent(
			"authorize_circuit_breaker",
			sdk.NewAttribute("granter", authority),
			sdk.NewAttribute("grantee", addresses[3]),
			sdk.NewAttribute("permission", somemsgs.String()),
		),
		lastEvent(ft.ctx),
	)

	add3, err := ft.ac.StringToBytes(addresses[3])
	require.NoError(t, err)

	perms, err = ft.keeper.Permissions.Get(ft.ctx, add3)
	require.NoError(t, err)

	require.Equal(t, somemsgs, perms)

	add4, err := ft.ac.StringToBytes(addresses[4])
	require.NoError(t, err)

	perms, err = ft.keeper.Permissions.Get(ft.ctx, add4)
	require.ErrorIs(t, err, collections.ErrNotFound, "users have no perms by default")

	require.Equal(t, types.Permissions{Level: types.Permissions_LEVEL_NONE_UNSPECIFIED, LimitTypeUrls: nil}, perms, "users have no perms by default")

	// admin tries grants another user permission SOME_MSGS with limited urls populated
	permis := types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{msgSend}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[4], Permissions: &permis}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)
}

func TestTripCircuitBreaker(t *testing.T) {
	ft := initFixture(t)

	srv := keeper.NewMsgServerImpl(ft.keeper)
	url := msgSend

	authority, err := ft.ac.BytesToString(ft.mockAddr)
	require.NoError(t, err)

	// admin trips circuit breaker
	admintrip := &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewEvent(
			"trip_circuit_breaker",
			sdk.NewAttribute("authority", authority),
			sdk.NewAttribute("msg_url", url),
		),
		lastEvent(ft.ctx),
	)

	allowed, err := ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	// user with enough permissions tries to trip circuit breaker for two messages
	url, url2 := "cosmos.gov.v1beta1.MsgDeposit", "cosmos.gov.v1beta1.MsgVote"
	twomsgs := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{url, url2}}
	msg := &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[3], Permissions: twomsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	// try to trip two messages with enough permissions
	twoMsgTrip := &types.MsgTripCircuitBreaker{Authority: addresses[3], MsgTypeUrls: []string{url, url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, twoMsgTrip)
	require.NoError(t, err)

	// user with all messages trips circuit breaker
	// add a super user
	allmsgs := &types.Permissions{Level: types.Permissions_LEVEL_ALL_MSGS, LimitTypeUrls: []string{""}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[1], Permissions: allmsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	// try to trip the circuit breaker
	url2 = "cosmos.staking.v1beta1.MsgDelegate"
	superTrip := &types.MsgTripCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, superTrip)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewEvent(
			"trip_circuit_breaker",
			sdk.NewAttribute("authority", addresses[1]),
			sdk.NewAttribute("msg_url", url2),
		),
		lastEvent(ft.ctx),
	)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url2)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	// user with no permission attempts to trip circuit breaker
	unknownTrip := &types.MsgTripCircuitBreaker{Authority: addresses[4], MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, unknownTrip)
	require.Error(t, err)

	// user tries to trip circuit breaker for two messages but only has permission for one
	url, url2 = "cosmos.staking.v1beta1.MsgCreateValidator", "cosmos.staking.v1beta1.MsgEditValidator"
	somemsgs := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{url}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[2], Permissions: somemsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	// try to trip two messages but user only has permission for one
	someTrip := &types.MsgTripCircuitBreaker{Authority: addresses[2], MsgTypeUrls: []string{url, url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, someTrip)
	require.ErrorContains(t, err, "MsgEditValidator: unauthorized")

	// user tries to trip an already tripped circuit breaker
	alreadyTripped := msgSend
	twoTrip := &types.MsgTripCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{alreadyTripped}}
	_, err = srv.TripCircuitBreaker(ft.ctx, twoTrip)
	require.ErrorContains(t, err, "already disabled")
}

func TestResetCircuitBreaker(t *testing.T) {
	ft := initFixture(t)
	authority, err := ft.ac.BytesToString(ft.mockAddr)
	require.NoError(t, err)

	srv := keeper.NewMsgServerImpl(ft.keeper)

	// admin resets circuit breaker
	url := msgSend
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
	require.Equal(
		t,
		sdk.NewEvent(
			"reset_circuit_breaker",
			sdk.NewAttribute("authority", authority),
			sdk.NewAttribute("msg_url", url),
		),
		lastEvent(ft.ctx),
	)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.True(t, allowed, "circuit breaker should be reset")

	// admin trips circuit breaker
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	// user has no permission to reset circuit breaker
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
	require.Equal(
		t,
		sdk.NewEvent(
			"reset_circuit_breaker",
			sdk.NewAttribute("authority", addresses[1]),
			sdk.NewAttribute("msg_url", url),
		),
		lastEvent(ft.ctx),
	)

	// user tries to reset a message they dont have permission to reset
	url = "cosmos.staking.v1beta1.MsgCreateValidator"
	// give restricted perms to a user
	someMsgs := &types.Permissions{Level: types.Permissions_LEVEL_SOME_MSGS, LimitTypeUrls: []string{url2}}
	msg = &types.MsgAuthorizeCircuitBreaker{Granter: authority, Grantee: addresses[2], Permissions: someMsgs}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, msg)
	require.NoError(t, err)

	admintrip = &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	// user with some messages resets circuit breaker
	someMsgsReset := &types.MsgResetCircuitBreaker{Authority: addresses[2], MsgTypeUrls: []string{url2}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, someMsgsReset)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewEvent(
			"reset_circuit_breaker",
			sdk.NewAttribute("authority", addresses[2]),
			sdk.NewAttribute("msg_url", url2),
		),
		lastEvent(ft.ctx),
	)

	// user tries to reset an already reset circuit breaker
	someMsgsReset = &types.MsgResetCircuitBreaker{Authority: addresses[1], MsgTypeUrls: []string{url2}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, someMsgsReset)
	require.Error(t, err)
}

func lastEvent(ctx context.Context) sdk.Event {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	events := sdkCtx.EventManager().Events()

	return events[len(events)-1]
}

func TestResetCircuitBreakerSomeMsgs(t *testing.T) {
	ft := initFixture(t)
	authority, err := ft.ac.BytesToString(ft.mockAddr)
	require.NoError(t, err)

	srv := keeper.NewMsgServerImpl(ft.keeper)

	// admin resets circuit breaker
	url := msgSend
	url2 := "the_only_message_acc2_can_trip_and_reset"

	// add acc2 as an authorized account for only url2
	authmsg := &types.MsgAuthorizeCircuitBreaker{
		Granter: authority,
		Grantee: addresses[2],
		Permissions: &types.Permissions{
			Level:         types.Permissions_LEVEL_SOME_MSGS,
			LimitTypeUrls: []string{url2},
		},
	}
	_, err = srv.AuthorizeCircuitBreaker(ft.ctx, authmsg)
	require.NoError(t, err)

	// admin trips circuit breaker
	admintrip := &types.MsgTripCircuitBreaker{Authority: authority, MsgTypeUrls: []string{url, url2}}
	_, err = srv.TripCircuitBreaker(ft.ctx, admintrip)
	require.NoError(t, err)

	// sanity check, both messages should be tripped
	allowed, err := ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url2)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	// now let's try to reset url using acc2 (should fail)
	acc2Reset := &types.MsgResetCircuitBreaker{Authority: addresses[2], MsgTypeUrls: []string{url}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, acc2Reset)
	require.Error(t, err)

	// now let's try to reset url2 using acc2 (should pass)
	acc2Reset = &types.MsgResetCircuitBreaker{Authority: addresses[2], MsgTypeUrls: []string{url2}}
	_, err = srv.ResetCircuitBreaker(ft.ctx, acc2Reset)
	require.NoError(t, err)

	// Only url2 should be reset, url should still be tripped
	allowed, err = ft.keeper.IsAllowed(ft.ctx, url)
	require.NoError(t, err)
	require.False(t, allowed, "circuit breaker should be tripped")

	allowed, err = ft.keeper.IsAllowed(ft.ctx, url2)
	require.NoError(t, err)
	require.True(t, allowed, "circuit breaker should be reset")
}
