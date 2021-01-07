package changepubkey

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	changepubkeykeeper "github.com/cosmos/cosmos-sdk/x/auth/changepubkey/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
)

// NewHandler returns a handler for x/auth message types.
func NewHandler(ak authkeeper.AccountKeeper, hk changepubkeykeeper.Keeper) sdk.Handler {
	msgServer := changepubkeykeeper.NewMsgServerImpl(ak, hk)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgChangePubKey:
			res, err := msgServer.ChangePubKey(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}
