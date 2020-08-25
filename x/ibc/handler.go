package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	porttypes "github.com/cosmos/cosmos-sdk/x/ibc/05-port/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/keeper"
)

// NewHandler defines the IBC handler
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		// IBC client msg interface types
		case clientexported.MsgCreateClient:
			return client.HandleMsgCreateClient(ctx, k.ClientKeeper, msg)

		case clientexported.MsgUpdateClient:
			return client.HandleMsgUpdateClient(ctx, k.ClientKeeper, msg)

		// Client Misbehaviour is handled by the evidence module

		// IBC connection msgs
		case *connectiontypes.MsgConnectionOpenInit:
			return connection.HandleMsgConnectionOpenInit(ctx, k.ConnectionKeeper, msg)

		case *connectiontypes.MsgConnectionOpenTry:
			return connection.HandleMsgConnectionOpenTry(ctx, k.ConnectionKeeper, msg)

		case *connectiontypes.MsgConnectionOpenAck:
			return connection.HandleMsgConnectionOpenAck(ctx, k.ConnectionKeeper, msg)

		case *connectiontypes.MsgConnectionOpenConfirm:
			return connection.HandleMsgConnectionOpenConfirm(ctx, k.ConnectionKeeper, msg)

		// IBC channel msgs
		case *channeltypes.MsgChannelOpenInit:
			// Lookup module by port capability
			module, portCap, err := k.PortKeeper.LookupModuleByPort(ctx, msg.PortId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			res, cap, err := channel.HandleMsgChannelOpenInit(ctx, k.ChannelKeeper, portCap, msg)
			if err != nil {
				return nil, err
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			if err = cbs.OnChanOpenInit(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortId, msg.ChannelId, cap, msg.Channel.Counterparty, msg.Channel.Version); err != nil {
				return nil, sdkerrors.Wrap(err, "channel open init callback failed")
			}

			return res, nil

		case *channeltypes.MsgChannelOpenTry:
			// Lookup module by port capability
			module, portCap, err := k.PortKeeper.LookupModuleByPort(ctx, msg.PortId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			res, cap, err := channel.HandleMsgChannelOpenTry(ctx, k.ChannelKeeper, portCap, msg)
			if err != nil {
				return nil, err
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			if err = cbs.OnChanOpenTry(ctx, msg.Channel.Ordering, msg.Channel.ConnectionHops, msg.PortId, msg.ChannelId, cap, msg.Channel.Counterparty, msg.Channel.Version, msg.CounterpartyVersion); err != nil {
				return nil, sdkerrors.Wrap(err, "channel open try callback failed")
			}

			return res, nil

		case *channeltypes.MsgChannelOpenAck:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			if err = cbs.OnChanOpenAck(ctx, msg.PortId, msg.ChannelId, msg.CounterpartyVersion); err != nil {
				return nil, sdkerrors.Wrap(err, "channel open ack callback failed")
			}

			return channel.HandleMsgChannelOpenAck(ctx, k.ChannelKeeper, cap, msg)

		case *channeltypes.MsgChannelOpenConfirm:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			if err = cbs.OnChanOpenConfirm(ctx, msg.PortId, msg.ChannelId); err != nil {
				return nil, sdkerrors.Wrap(err, "channel open confirm callback failed")
			}

			return channel.HandleMsgChannelOpenConfirm(ctx, k.ChannelKeeper, cap, msg)

		case *channeltypes.MsgChannelCloseInit:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			if err = cbs.OnChanCloseInit(ctx, msg.PortId, msg.ChannelId); err != nil {
				return nil, sdkerrors.Wrap(err, "channel close init callback failed")
			}

			return channel.HandleMsgChannelCloseInit(ctx, k.ChannelKeeper, cap, msg)

		case *channeltypes.MsgChannelCloseConfirm:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.PortId, msg.ChannelId)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			if err = cbs.OnChanCloseConfirm(ctx, msg.PortId, msg.ChannelId); err != nil {
				return nil, sdkerrors.Wrap(err, "channel close confirm callback failed")
			}

			return channel.HandleMsgChannelCloseConfirm(ctx, k.ChannelKeeper, cap, msg)

		// IBC packet msgs get routed to the appropriate module callback
		case *channeltypes.MsgRecvPacket:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.DestinationPort, msg.Packet.DestinationChannel)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			// For now, convert uint64 heights to clientexported.Height
			proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
			// Perform TAO verification
			if err := k.ChannelKeeper.RecvPacket(ctx, msg.Packet, msg.Proof, proofHeight); err != nil {
				return nil, sdkerrors.Wrap(err, "receive packet verification failed")
			}

			// Perform application logic callback
			res, ack, err := cbs.OnRecvPacket(ctx, msg.Packet)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "receive packet callback failed")
			}

			// Set packet acknowledgement
			if err = k.ChannelKeeper.PacketExecuted(ctx, cap, msg.Packet, ack); err != nil {
				return nil, err
			}

			return res, nil

		case *channeltypes.MsgAcknowledgement:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.SourcePort, msg.Packet.SourceChannel)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			// For now, convert uint64 heights to clientexported.Height
			proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
			// Perform TAO verification
			if err := k.ChannelKeeper.AcknowledgePacket(ctx, msg.Packet, msg.Acknowledgement, msg.Proof, proofHeight); err != nil {
				return nil, sdkerrors.Wrap(err, "acknowledge packet verification failed")
			}

			// Perform application logic callback
			res, err := cbs.OnAcknowledgementPacket(ctx, msg.Packet, msg.Acknowledgement)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "acknowledge packet callback failed")
			}

			// Delete packet commitment
			if err = k.ChannelKeeper.AcknowledgementExecuted(ctx, cap, msg.Packet); err != nil {
				return nil, err
			}

			return res, nil

		case *channeltypes.MsgTimeout:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.SourcePort, msg.Packet.SourceChannel)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			// For now, convert uint64 heights to clientexported.Height
			proofHeight := clientexported.NewHeight(msg.ProofEpoch, msg.ProofHeight)
			// Perform TAO verification
			if err := k.ChannelKeeper.TimeoutPacket(ctx, msg.Packet, msg.Proof, proofHeight, msg.NextSequenceRecv); err != nil {
				return nil, sdkerrors.Wrap(err, "timeout packet verification failed")
			}

			// Perform application logic callback
			res, err := cbs.OnTimeoutPacket(ctx, msg.Packet)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "timeout packet callback failed")
			}

			// Delete packet commitment
			if err = k.ChannelKeeper.TimeoutExecuted(ctx, cap, msg.Packet); err != nil {
				return nil, err
			}

			return res, nil

		case *channeltypes.MsgTimeoutOnClose:
			// Lookup module by channel capability
			module, cap, err := k.ChannelKeeper.LookupModuleByChannel(ctx, msg.Packet.SourcePort, msg.Packet.SourceChannel)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "could not retrieve module from port-id")
			}

			// Retrieve callbacks from router
			cbs, ok := k.Router.GetRoute(module)
			if !ok {
				return nil, sdkerrors.Wrapf(porttypes.ErrInvalidRoute, "route not found to module: %s", module)
			}

			// Perform TAO verification
			if err := k.ChannelKeeper.TimeoutOnClose(ctx, cap, msg.Packet, msg.Proof, msg.ProofClose, msg.ProofHeight, msg.NextSequenceRecv); err != nil {
				return nil, sdkerrors.Wrap(err, "timeout on close packet verification failed")
			}

			// Perform application logic callback
			// NOTE: MsgTimeout and MsgTimeoutOnClose use the same "OnTimeoutPacket"
			// application logic callback.
			res, err := cbs.OnTimeoutPacket(ctx, msg.Packet)
			if err != nil {
				return nil, sdkerrors.Wrap(err, "timeout packet callback failed")
			}

			// Delete packet commitment
			if err = k.ChannelKeeper.TimeoutExecuted(ctx, cap, msg.Packet); err != nil {
				return nil, err
			}

			return res, nil

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized IBC message type: %T", msg)
		}
	}
}
