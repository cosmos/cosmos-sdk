package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// SendTransfer handles transfer sending logic.
// The amount MUST be prefixed if it
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	destHeight uint64,
	amount sdk.Coins,
	sender,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	// get the port and channel of the counterparty
	sourceChan, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrap(channel.ErrChannelNotFound, sourceChannel)
	}

	destinationPort := sourceChan.Counterparty.PortID
	destinationChannel := sourceChan.Counterparty.ChannelID

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channel.ErrSequenceSendNotFound
	}

	coins := make(sdk.Coins, len(amount))
	switch {
	case !isSourceChain:
		// build the receiving denomination prefix
		prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		for i, coin := range amount {
			if !strings.HasPrefix(coin.Denom, prefix) {
				coins[i] = sdk.NewCoin(prefix+coin.Denom, coin.Amount)
			}
		}
	default:
		coins = amount
	}

	return k.createOutgoingPacket(ctx, sequence, sourcePort, sourceChannel, destinationPort, destinationChannel, destHeight, coins, sender, receiver, isSourceChain)
}

// ReceiveTransfer handles transfer receiving logic.
func (k Keeper) ReceiveTransfer(ctx sdk.Context, packet channel.Packet) error {
	return k.onRecvPacket(ctx, packet)
}

// TimeoutTransfer handles transfer timeout logic.
func (k Keeper) TimeoutTransfer(ctx sdk.Context, packet channel.Packet) error {
	return k.onTimeoutPacket(ctx, packet)
}

// See spec for this function: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func (k Keeper) createOutgoingPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	destHeight uint64,
	amount sdk.Coins,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	if isSourceChain {
		// escrow tokens if the destination chain is the same as the sender's
		escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

		// NOTE: we can omit the destination prefix correctness since it's already populated
		// internally on the SendTransfer function.

		// escrow source tokens. It fails if balance insufficient.
		if err := k.bankKeeper.SendCoins(
			ctx, sender, escrowAddress, amount,
		); err != nil {
			return err
		}
	} else {
		// burn vouchers from the sender's balance if the source is from another chain

		// NOTE: we can omit the source prefix correctness since it's already populated
		// internally on the SendTransfer function.

		// transfer the coins to the module account and burn them
		if err := k.supplyKeeper.SendCoinsFromAccountToModule(
			ctx, sender, types.GetModuleAccountName(), amount,
		); err != nil {
			return err
		}

		// burn from supply
		if err := k.supplyKeeper.BurnCoins(
			ctx, types.GetModuleAccountName(), amount,
		); err != nil {
			// NOTE: should not happen as the module account was
			// retrieved on the step above and it has enough balace
			// to burn.
			return err
		}
	}

	packetData := types.NewFungibleTokenPacketData(
		amount, sender, receiver, isSourceChain, destHeight+DefaultPacketTimeout,
	)

	packet := channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)

	return k.channelKeeper.SendPacket(ctx, packet)
}

func (k Keeper) onRecvPacket(ctx sdk.Context, packet channel.Packet) error {
	data, ok := packet.GetData().(types.FungibleTokenPacketData)
	if !ok {
		return sdkerrors.Wrap(
			channel.ErrInvalidPacket,
			"data doesn't correspond to fungible token transfer",
		)
	}

	if data.Source {
		prefix := types.GetDenomPrefix(packet.GetDestChannel(), packet.GetDestChannel())
		for _, coin := range data.Amount {
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdkerrors.Wrapf(
					sdkerrors.ErrInvalidCoins,
					"%s doesn't contain the prefix '%s'", coin.Denom, prefix,
				)
			}
		}

		// mint new tokens if the source of the transfer is the same chain
		if err := k.supplyKeeper.MintCoins(
			ctx, types.GetModuleAccountName(), data.Amount,
		); err != nil {
			return err
		}

		// send to receiver
		return k.supplyKeeper.SendCoinsFromModuleToAccount(
			ctx, types.GetModuleAccountName(), data.Receiver, data.Amount,
		)
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidCoins,
				"%s doesn't contain the prefix '%s'", coin.Denom, prefix,
			)
		}
		coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	}

	// unescrow tokens
	escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
	return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Receiver, coins)
}

func (k Keeper) onTimeoutPacket(ctx sdk.Context, packet channel.Packet) error {
	data, ok := packet.GetData().(types.FungibleTokenPacketData)
	if !ok {
		return sdkerrors.Wrap(
			channel.ErrInvalidPacket,
			"data doesn't correspond to fungible token transfer",
		)
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		coin := coin
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "%s doesn't contain the prefix '%s'", coin.Denom, prefix)
		}
		coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	}

	if data.Source {
		escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
		return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Sender, coins)
	}

	// mint from supply
	if err := k.supplyKeeper.MintCoins(
		ctx, types.GetModuleAccountName(), data.Amount,
	); err != nil {
		return err
	}

	return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GetModuleAccountName(), data.Sender, data.Amount)
}
