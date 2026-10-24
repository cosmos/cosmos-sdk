package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/safety/types"
)

type msgServer struct {
	Keeper
}

var _ types.MsgServer = msgServer{}

// NewMsgServerImpl returns an implementation of the x/safety MsgServer.
func NewMsgServerImpl(k Keeper) types.MsgServer {
	return msgServer{Keeper: k}
}

// UpdateParams replaces the params. Only the authority may call it. It never
// changes the disabled list.
func (s msgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if err := msg.Validate(s.addressCodec); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := s.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if err := sdk.ValidateAuthority(sdkCtx, params.Authority, msg.Authority); err != nil {
		return nil, s.notCanonical(msg.Authority, err)
	}

	if err := s.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(&types.EventParamsUpdated{
		Authority: msg.Authority,
		Params:    msg.Params,
	}); err != nil {
		return nil, err
	}
	return &types.MsgUpdateParamsResponse{}, nil
}

// Trip disables the given patterns. The sender must be the configured
// operator and every pattern must be covered by its scope, block a registered
// unprotected msg type, and not already be in the disabled list. A pattern
// under a broader disabled entry is accepted, so the operator can pin a msg
// before that entry is reset. Either every pattern is disabled or none is.
func (s msgServer) Trip(ctx context.Context, msg *types.MsgTrip) (*types.MsgTripResponse, error) {
	if err := msg.Validate(s.addressCodec); err != nil {
		return nil, err
	}

	params, err := s.Params.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !isOperator(params, msg.Operator) {
		return nil, sdkerrors.ErrUnauthorized.Wrapf("%s is not the operator", msg.Operator)
	}
	for _, p := range msg.Patterns {
		if !types.CoveredByAny(params.Scope, p) {
			return nil, types.ErrOutOfScope.Wrap(p)
		}
		if !s.matchesUnprotectedMsg(p) {
			return nil, types.ErrNoMatchingMsg.Wrap(p)
		}
		has, err := s.DisabledList.Has(ctx, p)
		if err != nil {
			return nil, err
		}
		if has {
			return nil, types.ErrAlreadyDisabled.Wrap(p)
		}
	}

	for _, p := range msg.Patterns {
		if err := s.DisabledList.Set(ctx, p); err != nil {
			return nil, err
		}
	}

	if err := sdk.UnwrapSDKContext(ctx).EventManager().EmitTypedEvent(&types.EventTrip{
		Operator: msg.Operator,
		Patterns: msg.Patterns,
	}); err != nil {
		return nil, err
	}
	return &types.MsgTripResponse{}, nil
}

// Reset removes the given patterns from the disabled list. The authority may
// reset any pattern; the operator may reset patterns covered by its scope.
// Every pattern must be present in the disabled list exactly. Either every
// pattern is reset or none is.
func (s msgServer) Reset(ctx context.Context, msg *types.MsgReset) (*types.MsgResetResponse, error) {
	if err := msg.Validate(s.addressCodec); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	params, err := s.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	if authErr := sdk.ValidateAuthority(sdkCtx, params.Authority, msg.Sender); authErr != nil {
		if !isOperator(params, msg.Sender) {
			return nil, s.notCanonical(msg.Sender,
				sdkerrors.ErrUnauthorized.Wrapf("%s is neither the authority nor the operator", msg.Sender))
		}
		for _, p := range msg.Patterns {
			if !types.CoveredByAny(params.Scope, p) {
				return nil, types.ErrOutOfScope.Wrap(p)
			}
		}
	}

	for _, p := range msg.Patterns {
		has, err := s.DisabledList.Has(ctx, p)
		if err != nil {
			return nil, err
		}
		if !has {
			return nil, types.ErrNotDisabled.Wrap(p)
		}
	}

	for _, p := range msg.Patterns {
		if err := s.DisabledList.Remove(ctx, p); err != nil {
			return nil, err
		}
	}

	if err := sdkCtx.EventManager().EmitTypedEvent(&types.EventReset{
		Sender:   msg.Sender,
		Patterns: msg.Patterns,
	}); err != nil {
		return nil, err
	}
	return &types.MsgResetResponse{}, nil
}

// notCanonical returns ErrInvalidAddress when a sender that matched no role
// is not in canonical form, and err otherwise. Roles match by exact string,
// as sdk.ValidateAuthority does, so the same account in another case is not
// the authority or the operator, and saying so is clearer than unauthorized.
// The reset and UpdateParams senders are not required to be canonical up
// front, because a consensus authority override may itself be non-canonical.
func (s msgServer) notCanonical(sender string, err error) error {
	if cerr := types.ValidateCanonicalAddress(s.addressCodec, sender); cerr != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("%s: %s", sender, cerr)
	}
	return err
}

// isOperator reports whether addr is the configured, non-empty operator.
// Params validation requires a canonical operator, so the strings compare
// directly.
func isOperator(params types.Params, addr string) bool {
	return params.Operator != "" && params.Operator == addr
}
