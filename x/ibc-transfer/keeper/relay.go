package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// SendTransfer handles transfer sending logic. There are 2 possible cases:
//
// 1. Sender chain is acting as the source zone. The coins are transferred to an
// escrow address (i.e locked) on the sender chain and then transferred to the
// receiving chain through IBC TAO logic. It is expected that the receiving
// chain will mint vouchers to the receiving address.
//
// 2. Sender chain is acting as the sink zone. The coins (vouchers) are burned
// on the sender chain and then transferred to the receiving chain though IBC
// TAO logic. It is expected that the receiving chain will unescrow the fungible
// token and send it to the receiving address.
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	token sdk.Coin,
	sender sdk.AccAddress,
	receiver string,
	source bool,
	timeoutHeight,
	timeoutTimestamp uint64,
) error {
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(channeltypes.ErrChannelNotFound, "port ID (%s) channel ID (%s)", sourcePort, sourceChannel)
	}

	destinationPort := sourceChannelEnd.GetCounterparty().GetPortID()
	destinationChannel := sourceChannelEnd.GetCounterparty().GetChannelID()

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrapf(
			channeltypes.ErrSequenceSendNotFound,
			"source port: %s, source channel: %s", sourcePort, sourceChannel,
		)
	}

	// begin createOutgoingPacket logic
	// See spec for this logic: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay

	channelCap, ok := k.scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePort, sourceChannel))
	if !ok {
		return sdkerrors.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}
	// NOTE: When the sender chain is acting as the source zone, the coins will
	// be constructed with the destination port and channel ID prefixed when
	// represented in the fungible token packet data. However the coins
	// will be escrowed on the source chain without the prefix

	if source {
		// create the escrow address for the tokens
		escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

		// escrow source tokens. It fails if balance insufficient.
		if err := k.bankKeeper.SendCoins(
			ctx, sender, escrowAddress, sdk.NewCoins(token),
		); err != nil {
			return err
		}

		// construct denom with prefix that will be used in the transfer
		token.Denom = types.GetPrefixedDenom(destinationPort, destinationChannel, token.Denom)

	} else {
		// ensure that the coin has the correct prefix
		prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		if !strings.HasPrefix(token.Denom, prefix) {
			return sdkerrors.Wrapf(
				types.ErrInvalidDenomForTransfer,
				"%s doesn't contain the prefix '%s'", token.Denom, prefix,
			)
		}

		// transfer the coins to the module account and burn them
		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, sender, types.ModuleName, sdk.NewCoins(token),
		); err != nil {
			return err
		}

		if err := k.bankKeeper.BurnCoins(
			ctx, types.ModuleName, sdk.NewCoins(token),
		); err != nil {
			// NOTE: should not happen as the module account was
			// retrieved on the step above and it has enough balace
			// to burn.
			return err
		}
	}

	packetData := types.NewFungibleTokenPacketData(
		token.Denom, token.Amount.Uint64(), sender.String(), receiver, source,
	)

	packet := channeltypes.NewPacket(
		packetData.GetBytes(),
		sequence,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		timeoutHeight,
		timeoutTimestamp,
	)

	return k.channelKeeper.SendPacket(ctx, channelCap, packet)
}

func (k Keeper) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	// NOTE: packet data type already checked in handler.go
	token := sdk.NewCoin(data.Denom, sdk.NewIntFromUint64(data.Amount))

	// decode the receiver address
	receiver, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return err
	}

	if data.Source {
		// ensure that the coin has the correct prefix
		prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		if !strings.HasPrefix(token.Denom, prefix) {
			return sdkerrors.Wrapf(
				types.ErrInvalidDenomForTransfer,
				"%s doesn't contain the prefix '%s'", token.Denom, prefix,
			)
		}

		// mint new tokens if the source of the transfer is the same chain
		if err := k.bankKeeper.MintCoins(
			ctx, types.ModuleName, sdk.NewCoins(token),
		); err != nil {
			return err
		}

		// send to receiver
		return k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx, types.ModuleName, receiver, sdk.NewCoins(token),
		)
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	if !strings.HasPrefix(token.Denom, prefix) {
		return sdkerrors.Wrapf(
			types.ErrInvalidDenomForTransfer,
			"%s doesn't contain the prefix '%s'", token.Denom, prefix,
		)
	}
	token.Denom = token.Denom[len(prefix):]

	// unescrow tokens
	escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
	return k.bankKeeper.SendCoins(ctx, escrowAddress, receiver, sdk.NewCoins(token))
}

func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData, ack types.FungibleTokenPacketAcknowledgement) error {
	if !ack.Success {
		return k.refundPacketToken(ctx, packet, data)
	}
	return nil
}

func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	return k.refundPacketToken(ctx, packet, data)
}

func (k Keeper) refundPacketToken(ctx sdk.Context, packet channeltypes.Packet, data types.FungibleTokenPacketData) error {
	// NOTE: packet data type already checked in handler.go

	token := sdk.NewCoin(data.Denom, sdk.NewIntFromUint64(data.Amount))

	// decode the sender address
	sender, err := sdk.AccAddressFromBech32(data.Sender)
	if err != nil {
		return err
	}

	if data.Source {
		// check the denom prefix
		prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		if !strings.HasPrefix(token.Denom, prefix) {
			return sdkerrors.Wrapf(
				types.ErrInvalidDenomForTransfer,
				"%s doesn't contain the prefix '%s'", token.Denom, prefix,
			)
		}
		token.Denom = token.Denom[len(prefix):]

		// unescrow tokens back to sender
		escrowAddress := types.GetEscrowAddress(packet.GetSourcePort(), packet.GetSourceChannel())
		return k.bankKeeper.SendCoins(ctx, escrowAddress, sender, sdk.NewCoins(token))
	}

	// mint vouchers back to sender
	if err := k.bankKeeper.MintCoins(
		ctx, types.ModuleName, sdk.NewCoins(token),
	); err != nil {
		return err
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sender, sdk.NewCoins(token))
}
