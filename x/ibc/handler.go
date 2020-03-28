package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// NewHandler defines the IBC handler
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// IBC client msg interface types
		case clientexported.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, k.ClientKeeper, msg)

		case clientexported.MsgUpdateClient:
			return &sdk.Result{}, nil

		// IBC connection  msgs
		case connection.MsgConnectionOpenInit:
			return connection.HandleMsgConnectionOpenInit(ctx, k.ConnectionKeeper, msg)

		case connection.MsgConnectionOpenTry:
			return connection.HandleMsgConnectionOpenTry(ctx, k.ConnectionKeeper, msg)

		case connection.MsgConnectionOpenAck:
			return connection.HandleMsgConnectionOpenAck(ctx, k.ConnectionKeeper, msg)

		case connection.MsgConnectionOpenConfirm:
			return connection.HandleMsgConnectionOpenConfirm(ctx, k.ConnectionKeeper, msg)

		// IBC channel msgs
		case channel.MsgChannelOpenInit:
			cbs := k.PortKeeper.LookupModule(ctx, msg.PortID)
			portCap, _ := cbs.GetCapability(ctx, ibctypes.PortPath(msg.PortID))
			err := cbs.OnChanOpenInit(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID, msg.Channel.Counterparty, msg.Channel.Version)
			if err != nil {
				return nil, err
			}
			res, cap, err := channel.HandleMsgChannelOpenInit(ctx, k.ChannelKeeper, portCap, msg)
			cbs.ClaimCapability(ctx, cap, ibctypes.ChannelCapabilityPath(msg.PortID, msg.ChannelID))
			return res, err

		case channel.MsgChannelOpenTry:
			cbs := k.PortKeeper.LookupModule(ctx, msg.PortID)
			portCap, _ := cbs.GetCapability(ctx, ibctypes.PortPath(msg.PortID))
			err := cbs.OnChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortID, msg.ChannelID, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion)
			if err != nil {
				return nil, err
			}
			res, cap, err := channel.HandleMsgChannelOpenTry(ctx, k.ChannelKeeper, portCap, msg)
			cbs.ClaimCapability(ctx, cap, ibctypes.ChannelCapabilityPath(msg.PortID, msg.ChannelID))
			return res, err

		case channel.MsgChannelOpenAck:
			cbs := k.PortKeeper.LookupModule(ctx, msg.PortID)
			chanCap, _ := cbs.GetCapability(ctx, ibctypes.ChannelCapabilityPath(msg.PortID, msg.ChannelID))
			err := cbs.OnChanOpenAck(ctx, msg.PortID, msg.ChannelID, msg.CounterpartyVersion)
			if err != nil {
				return nil, err
			}
			return channel.HandleMsgChannelOpenAck(ctx, k.ChannelKeeper, chanCap, msg)

		case channel.MsgChannelOpenConfirm:
			cbs := k.PortKeeper.LookupModule(ctx, msg.PortID)
			chanCap, _ := cbs.GetCapability(ctx, ibctypes.ChannelCapabilityPath(msg.PortID, msg.ChannelID))
			err := cbs.OnChanOpenConfirm(ctx, msg.PortID, msg.ChannelID)
			if err != nil {
				return nil, err
			}
			return channel.HandleMsgChannelOpenConfirm(ctx, k.ChannelKeeper, chanCap, msg)

		case channel.MsgChannelCloseInit:
			cbs := k.PortKeeper.LookupModule(ctx, msg.PortID)
			chanCap, _ := cbs.GetCapability(ctx, ibctypes.ChannelCapabilityPath(msg.PortID, msg.ChannelID))
			err := cbs.OnChanCloseInit(ctx, msg.PortID, msg.ChannelID)
			if err != nil {
				return nil, err
			}
			return channel.HandleMsgChannelCloseInit(ctx, k.ChannelKeeper, chanCap, msg)

		case channel.MsgChannelCloseConfirm:
			cbs := k.PortKeeper.LookupModule(ctx, msg.PortID)
			chanCap, _ := cbs.GetCapability(ctx, ibctypes.ChannelCapabilityPath(msg.PortID, msg.ChannelID))
			err := cbs.OnChanCloseConfirm(ctx, msg.PortID, msg.ChannelID)
			if err != nil {
				return nil, err
			}
			return channel.HandleMsgChannelCloseConfirm(ctx, k.ChannelKeeper, chanCap, msg)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}
