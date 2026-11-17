package keeper

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

type msgServer struct {
	Keeper
}

var _ authz.MsgServer = msgServer{}

// newMsgServerImpl returns an implementation of the x/authz MsgServer
// interface for the provided Keeper.
func newMsgServerImpl(k Keeper) authz.MsgServer {
	return msgServer{Keeper: k}
}

// RegisterMsgServer registers the x/authz Msg service on r. The msg server
// is unexported: code outside the module runs a msg through the
// MsgServiceRouter with msgservice.Dispatch.
func RegisterMsgServer(r grpc.ServiceRegistrar, k Keeper) {
	authz.RegisterMsgServer(r, newMsgServerImpl(k))
}

// Grant implements the MsgServer.Grant method to create a new grant.
func (k msgServer) Grant(goCtx context.Context, msg *authz.MsgGrant) (*authz.MsgGrantResponse, error) {
	if strings.EqualFold(msg.Grantee, msg.Granter) {
		return nil, authz.ErrGranteeIsGranter
	}

	grantee, err := k.authKeeper.AddressCodec().StringToBytes(msg.Grantee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid grantee address: %s", err)
	}

	granter, err := k.authKeeper.AddressCodec().StringToBytes(msg.Granter)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid granter address: %s", err)
	}

	if err := msg.Grant.ValidateBasic(); err != nil {
		return nil, err
	}

	// create the account if it is not in account state
	ctx := sdk.UnwrapSDKContext(goCtx)
	granteeAcc := k.authKeeper.GetAccount(ctx, grantee)
	if granteeAcc == nil {
		if k.bankKeeper.BlockedAddr(grantee) {
			return nil, sdkerrors.ErrUnauthorized.Wrapf("%s is not allowed to receive funds", grantee)
		}

		granteeAcc = k.authKeeper.NewAccountWithAddress(ctx, grantee)
		k.authKeeper.SetAccount(ctx, granteeAcc)
	}

	authorization, err := msg.GetAuthorization()
	if err != nil {
		return nil, err
	}

	t := authorization.MsgTypeURL()
	if k.router.HandlerByTypeURL(t) == nil {
		return nil, sdkerrors.ErrInvalidType.Wrapf("%s doesn't exist.", t)
	}

	err = k.SaveGrant(ctx, grantee, granter, authorization, msg.Grant.Expiration)
	if err != nil {
		return nil, err
	}

	return &authz.MsgGrantResponse{}, nil
}

// Revoke implements the MsgServer.Revoke method.
func (k msgServer) Revoke(goCtx context.Context, msg *authz.MsgRevoke) (*authz.MsgRevokeResponse, error) {
	if strings.EqualFold(msg.Grantee, msg.Granter) {
		return nil, authz.ErrGranteeIsGranter
	}

	grantee, err := k.authKeeper.AddressCodec().StringToBytes(msg.Grantee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid grantee address: %s", err)
	}

	granter, err := k.authKeeper.AddressCodec().StringToBytes(msg.Granter)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid granter address: %s", err)
	}

	if msg.MsgTypeUrl == "" {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("missing msg method name")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err = k.DeleteGrant(ctx, grantee, granter, msg.MsgTypeUrl); err != nil {
		return nil, err
	}

	return &authz.MsgRevokeResponse{}, nil
}

// Exec implements the MsgServer.Exec method.
func (k msgServer) Exec(goCtx context.Context, msg *authz.MsgExec) (*authz.MsgExecResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if msg.Grantee == "" {
		return nil, errors.New("empty address string is not allowed")
	}

	grantee, err := k.authKeeper.AddressCodec().StringToBytes(msg.Grantee)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid grantee address: %s", err)
	}

	if len(msg.Msgs) == 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("messages cannot be empty")
	}

	msgs, err := msg.GetMessages()
	if err != nil {
		return nil, err
	}

	if err := validateMsgs(msgs); err != nil {
		return nil, err
	}

	results, err := k.DispatchActions(ctx, grantee, msgs)
	if err != nil {
		return nil, err
	}

	return &authz.MsgExecResponse{Results: results}, nil
}

func validateMsgs(msgs []sdk.Msg) error {
	for i, msg := range msgs {
		m, ok := msg.(sdk.HasValidateBasic)
		if !ok {
			continue
		}

		if err := m.ValidateBasic(); err != nil {
			return errorsmod.Wrapf(err, "msg %d", i)
		}
	}

	return nil
}

// Grant runs MsgGrant on the module's msg server.
//
// Deprecated: run MsgGrant with msgservice.Dispatch. Keeper.Grant will be
// removed in a future release.
func (k Keeper) Grant(ctx context.Context, msg *authz.MsgGrant) (*authz.MsgGrantResponse, error) {
	return newMsgServerImpl(k).Grant(ctx, msg)
}

// Revoke runs MsgRevoke on the module's msg server.
//
// Deprecated: run MsgRevoke with msgservice.Dispatch. Keeper.Revoke will be
// removed in a future release.
func (k Keeper) Revoke(ctx context.Context, msg *authz.MsgRevoke) (*authz.MsgRevokeResponse, error) {
	return newMsgServerImpl(k).Revoke(ctx, msg)
}

// Exec runs MsgExec on the module's msg server.
//
// Deprecated: run MsgExec with msgservice.Dispatch. Keeper.Exec will be
// removed in a future release.
func (k Keeper) Exec(ctx context.Context, msg *authz.MsgExec) (*authz.MsgExecResponse, error) {
	return newMsgServerImpl(k).Exec(ctx, msg)
}
