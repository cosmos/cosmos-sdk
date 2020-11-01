package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/keeper"
)

// NewHandler defines the IBC handler
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// IBC client msg interface types
		case *clienttypes.MsgCreateClient:
			res, err := k.CreateClient(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *clienttypes.MsgUpdateClient:
			res, err := k.UpdateClient(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *clienttypes.MsgUpgradeClient:
			res, err := k.UpgradeClient(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *clienttypes.MsgSubmitMisbehaviour:
			res, err := k.SubmitMisbehaviour(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		// IBC connection msgs
		case *connectiontypes.MsgConnectionOpenInit:
			res, err := k.ConnectionOpenInit(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *connectiontypes.MsgConnectionOpenTry:
			res, err := k.ConnectionOpenTry(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *connectiontypes.MsgConnectionOpenAck:
			res, err := k.ConnectionOpenAck(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *connectiontypes.MsgConnectionOpenConfirm:
			res, err := k.ConnectionOpenConfirm(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		// IBC channel msgs
		case *channeltypes.MsgChannelOpenInit:
			res, err := k.ChannelOpenInit(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgChannelOpenTry:
			res, err := k.ChannelOpenTry(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgChannelOpenAck:
			res, err := k.ChannelOpenAck(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgChannelOpenConfirm:
			res, err := k.ChannelOpenConfirm(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgChannelCloseInit:
			res, err := k.ChannelCloseInit(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgChannelCloseConfirm:
			res, err := k.ChannelCloseConfirm(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		// IBC packet msgs get routed to the appropriate module callback
		case *channeltypes.MsgRecvPacket:
			res, err := k.RecvPacket(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgAcknowledgement:
			res, err := k.Acknowledgement(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgTimeout:
			res, err := k.Timeout(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *channeltypes.MsgTimeoutOnClose:
			res, err := k.TimeoutOnClose(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}
