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
	perms, err := srv.GetPermissions(ctx, []byte(msg.Granter))
	if err != nil {
		return nil, err
	}
	if perms.Level != types.Permissions_LEVEL_SUPER_ADMIN {
		return nil, fmt.Errorf("only super admins can authorize circuit breakers")
	}

	// Append the account in the msg to the store's set of authorized super admins
	err = srv.SetPermissions(ctx, []byte(msg.Grantee), types.Permissions{
		Level: types.Permissions_LEVEL_SUPER_ADMIN,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgAuthorizeCircuitBreakerResponse{
		Success: true,
	}, nil
}

func (srv msgServer) TripCircuitBreaker(goCtx context.Context, msg *types.MsgTripCircuitBreaker) (*types.MsgTripCircuitBreakerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check that the account has the permission level of "super admin" or "circuit breaker"
	perms, err := srv.GetPermissions(ctx, []byte(msg.Authority))
	if err != nil {
		return nil, err
	}
	if perms.Level != types.Permissions_LEVEL_SUPER_ADMIN {
		return nil, fmt.Errorf("account does not have permission to trip circuit breaker")
	}

	store := ctx.KVStore(srv.key)
	// if the msg_type_urls is empty, add all msg type urls to the disable list
	if len(msg.MsgTypeUrls) == 0 {
		store.Set(types.DisableListPrefix, []byte{0x01})
	} else {
		// otherwise add the specific msg type urls to the disable list
		for _, msgTypeUrl := range msg.MsgTypeUrls {
			store.Set(append(types.DisableListPrefix, []byte(msgTypeUrl)...), []byte{0x01})
		}
	}

	return &types.MsgTripCircuitBreakerResponse{
		Success: true,
	}, nil
}

// ResetCircuitBreaker resumes processing of Msg's in the state machine that
// have been been paused using TripCircuitBreaker.

func (srv msgServer) ResetCircuitBreaker(ctx context.Context, msg *types.MsgResetCircuitBreaker) (*types.MsgResetCircuitBreakerResponse, error) {
	return nil, nil
	//check that the account has any permission and remove the typeurl from the disable list
}
