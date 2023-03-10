package keeper

import (
	"bytes"
	context "context"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the bank MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (srv msgServer) AuthorizeCircuitBreaker(goCtx context.Context, msg *types.MsgAuthorizeCircuitBreaker) (*types.MsgAuthorizeCircuitBreakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := srv.addressCodec.StringToBytes(msg.Granter)
	if err != nil {
		return nil, err
	}

	// if the granter is the module authority no need to check perms
	if !bytes.Equal(address, srv.GetAuthority()) {
		// Check that the authorizer has the permission level of "super admin"
		perms, err := srv.GetPermissions(ctx, address)
		if err != nil {
			return nil, err
		}

		if perms.Level != types.Permissions_LEVEL_SUPER_ADMIN {
			return nil, fmt.Errorf("only super admins can authorize users")
		}
	}

	grantee, err := srv.addressCodec.StringToBytes(msg.Grantee)
	if err != nil {
		return nil, err
	}

	// Append the account in the msg to the store's set of authorized super admins
	if err = srv.SetPermissions(ctx, grantee, msg.Permissions); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"authorize_circuit_breaker",
			sdk.NewAttribute("granter", msg.Granter),
			sdk.NewAttribute("grantee", msg.Grantee),
			sdk.NewAttribute("permission", msg.Permissions.String()),
		),
	})

	return &types.MsgAuthorizeCircuitBreakerResponse{
		Success: true,
	}, nil
}

func (srv msgServer) TripCircuitBreaker(goCtx context.Context, msg *types.MsgTripCircuitBreaker) (*types.MsgTripCircuitBreakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := srv.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(srv.storekey)

	// Check that the account has the permissions
	perms, err := srv.GetPermissions(ctx, address)
	if err != nil {
		return nil, err
	}

	switch {
	case perms.Level == types.Permissions_LEVEL_SUPER_ADMIN || bytes.Equal(address, srv.GetAuthority()):
		// add all msg type urls to the disable list
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			if !srv.IsAllowed(ctx, msgTypeUrl) {
				return nil, fmt.Errorf("message %s is already disabled", msgTypeUrl)
			}
			store.Set(types.CreateDisableMsgPrefix(msgTypeUrl), []byte{0x01})
		}
	case perms.Level == types.Permissions_LEVEL_ALL_MSGS:
		// iterate over the msg type urls
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			// check if the message is in the list of allowed messages
			if !srv.IsAllowed(ctx, msgTypeUrl) {
				return nil, fmt.Errorf("message %s is already disabled", msgTypeUrl)
			}
			store.Set(types.CreateDisableMsgPrefix(msgTypeUrl), []byte{0x01})
		}
	case perms.Level == types.Permissions_LEVEL_SOME_MSGS:
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			// check if the message is in the list of allowed messages
			if !srv.IsAllowed(ctx, msgTypeUrl) {
				return nil, fmt.Errorf("message %s is already disabled", msgTypeUrl)
			}
			for _, msgurl := range perms.LimitTypeUrls {
				if msgTypeUrl == msgurl {
					store.Set(types.CreateDisableMsgPrefix(msgTypeUrl), []byte{0x01})
				} else {
					return nil, fmt.Errorf("account does not have permission to trip circuit breaker for message %s", msgTypeUrl)
				}
			}
		}
	default:
		return nil, fmt.Errorf("account does not have permission to trip circuit breaker")
	}

	var msg_urls string
	if len(msg.GetMsgTypeUrls()) > 1 {

		for _, url := range msg.GetMsgTypeUrls() {
			msg_urls = msg_urls + ", " + url
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"trip_circuit_breaker",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("msg_url", msg_urls),
		),
	})

	return &types.MsgTripCircuitBreakerResponse{
		Success: true,
	}, nil
}

// ResetCircuitBreaker resumes processing of Msg's in the state machine that
// have been been paused using TripCircuitBreaker.
func (srv msgServer) ResetCircuitBreaker(goCtx context.Context, msg *types.MsgResetCircuitBreaker) (*types.MsgResetCircuitBreakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	keeper := srv.Keeper

	address, err := srv.addressCodec.StringToBytes(msg.Authority)
	if err != nil {
		return nil, err
	}

	// Get the permissions for the account specified in the msg.Authority field
	perms, err := keeper.GetPermissions(ctx, address)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(srv.storekey)

	switch {
	case perms.Level == types.Permissions_LEVEL_SUPER_ADMIN || bytes.Equal(address, srv.GetAuthority()):
		// add all msg type urls to the disable list
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			if srv.IsAllowed(ctx, msgTypeUrl) {
				return nil, fmt.Errorf("message %s is not disabled", msgTypeUrl)
			}
			store.Delete(types.CreateDisableMsgPrefix(msgTypeUrl))
		}
	case perms.Level == types.Permissions_LEVEL_ALL_MSGS:
		// iterate over the msg type urls
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			// check if the message is in the list of allowed messages
			if srv.IsAllowed(ctx, msgTypeUrl) {
				return nil, fmt.Errorf("message %s is not disabled", msgTypeUrl)
			}
			store.Delete(types.CreateDisableMsgPrefix(msgTypeUrl))
		}
	case perms.Level == types.Permissions_LEVEL_SOME_MSGS:
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			// check if the message is in the list of allowed messages
			if srv.IsAllowed(ctx, msgTypeUrl) {
				return nil, fmt.Errorf("message %s is not disabled", msgTypeUrl)
			}
			// allow user with limited permissions to reset circuit breaker for any message
			store.Delete(types.CreateDisableMsgPrefix(msgTypeUrl))

		}
	default:
		return nil, fmt.Errorf("account does not have permission to reset circuit breaker")
	}

	var msg_urls string
	if len(msg.GetMsgTypeUrls()) > 1 {

		for _, url := range msg.GetMsgTypeUrls() {
			msg_urls = msg_urls + ", " + url
		}
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			"reset_circuit_breaker",
			sdk.NewAttribute("authority", msg.Authority),
			sdk.NewAttribute("msg_url", msg_urls),
		),
	})

	return &types.MsgResetCircuitBreakerResponse{Success: true}, nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
