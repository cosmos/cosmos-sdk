package feegrant

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *MsgGrantFeeAllowance:
			return handleGrantFee(ctx, k, msg)

		case *MsgRevokeFeeAllowance:
			return handleRevokeFee(ctx, k, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

func handleGrantFee(ctx sdk.Context, k Keeper, msg *types.MsgGrantFeeAllowance) (*sdk.Result, error) {
	grantee, err := sdk.AccAddressFromBech32(msg.Grantee.String())
	if err != nil {
		return nil, err
	}

	granter, err := sdk.AccAddressFromBech32(msg.Granter.String())
	if err != nil {
		return nil, err
	}

	feegrant := types.FeeAllowanceGrant(types.FeeAllowanceGrant{
		Grantee:   grantee,
		Granter:   granter,
		Allowance: msg.Allowance,
	})

	k.GrantFeeAllowance(ctx, feegrant)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}

func handleRevokeFee(ctx sdk.Context, k Keeper, msg *types.MsgRevokeFeeAllowance) (*sdk.Result, error) {
	k.RevokeFeeAllowance(ctx, msg.Granter, msg.Grantee)
	return &sdk.Result{Events: ctx.EventManager().ABCIEvents()}, nil
}
