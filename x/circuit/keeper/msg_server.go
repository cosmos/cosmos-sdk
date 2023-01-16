package keeper

import (
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

	// Check that the authorizer has the permission level of "super admin"
	perms, err := srv.GetPermissions(ctx, msg.Granter)
	if err != nil {
		return nil, err
	}

	if perms.Level != types.Permissions_LEVEL_SUPER_ADMIN {
		return nil, fmt.Errorf("only super admins can authorize circuit breakers")
	}

	// Append the account in the msg to the store's set of authorized super admins
	if err = srv.SetPermissions(ctx, []byte(msg.Grantee), msg.Permissions); err != nil {
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

	// Check that the account has the permissions
	perms, err := srv.GetPermissions(ctx, msg.Authority)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(srv.key)

	switch perms.Level {
    case types.Permissions_LEVEL_SUPER_ADMIN:
        // add all msg type urls to the disable list
        for _, msgTypeUrl := range msg.MsgTypeUrls {
            store.Set(types.CreateDisableMsgPrefix(msgTypeUrl), []byte{0x01})
        }
    case types.Permissions_LEVEL_SOME_MSGS:
        // iterate over the msg type urls
        for _, msgTypeUrl := range msg.MsgTypeUrls {
            // check if the message is in the list of allowed messages
            if !store.Has(types.CreateDisableMsgPrefix(msgTypeUrl)) {
                return nil, fmt.Errorf("account does not have permission to trip circuit breaker for message %s", msgTypeUrl)
            }
            store.Set(types.CreateDisableMsgPrefix(msgTypeUrl), []byte{0x01})
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

	// Get the permissions for the account specified in the msg.Authority field
	accountPerms, err := keeper.GetPermissions(ctx, msg.Authority)
	if err != nil {
		return nil, err
	}

	// check that the account has permission to reset the circuit breaker
	if accountPerms.Level != types.Permissions_LEVEL_ALL_MSGS && accountPerms.Level != types.Permissions_LEVEL_SOME_MSGS {
		return nil, fmt.Errorf("account does not have permission to reset the circuit breaker")
	}

	// check that the account has permission to reset the specific msgTypeURLs, if any were provided
	if accountPerms.Level == types.Permissions_LEVEL_SOME_MSGS {
		for _, msgTypeURL := range msg.MsgTypeUrls {
			if !stringInSlice(msgTypeURL, accountPerms.LimitTypeUrls) {
				return nil, fmt.Errorf("account does not have permission to reset the specific msgTypeURL %s", msgTypeURL)
			}
			// Remove the type URL from the disable list
			keeper.EnableMsg(ctx, msgTypeURL)
		}
	} else if accountPerms.Level == types.Permissions_LEVEL_ALL_MSGS {
		for _, msgTypeURL := range msg.MsgTypeUrls {
			if !keeper.IsMsgDisabled(ctx, msgTypeURL) {
				return nil, fmt.Errorf("msgTypeURL %s is not disabled", msgTypeURL)
			}
			// Remove the type URL from the disable list
			keeper.EnableMsg(ctx, msgTypeURL)
		}
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
