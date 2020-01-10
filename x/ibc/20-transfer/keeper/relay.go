package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// SendTransfer handles transfer sending logic
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
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
	prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
	switch {
	case isSourceChain:
		// build the receiving denomination prefix
		for i, coin := range amount {
			coins[i] = sdk.NewCoin(prefix+coin.Denom, coin.Amount)
		}
	default:
		coins = amount
	}

	return k.createOutgoingPacket(ctx, sequence, sourcePort, sourceChannel, destinationPort, destinationChannel, coins, sender, receiver, isSourceChain)
}

// ReceiveTransfer handles transfer receiving logic
func (k Keeper) ReceiveTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataTransfer,
) error {
	if data.Source {
		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		for _, coin := range data.Amount {
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdkerrors.Wrapf(
					sdkerrors.ErrInvalidCoins,
					"%s doesn't contain the prefix '%s'", coin.Denom, prefix,
				)
			}
		}

		// mint new tokens if the source of the transfer is the same chain
		err := k.supplyKeeper.MintCoins(ctx, types.GetModuleAccountName(), data.Amount)
		if err != nil {
			return err
		}

		// send to receiver
		return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GetModuleAccountName(), data.Receiver, data.Amount)
	}

	// unescrow tokens

	// check the denom prefix
	prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
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

	escrowAddress := types.GetEscrowAddress(destinationPort, destinationChannel)
	return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Receiver, coins)

}

// TimeoutTransfer handles transfer timeout logic
func (k Keeper) TimeoutTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	data types.PacketDataTransfer,
) error {
	// check the denom prefix
	prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		coin := coin
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "%s doesn't contain the prefix '%s'", coin.Denom, prefix)
		}
		coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	}

	if data.Source {
		escrowAddress := types.GetEscrowAddress(destinationPort, destinationChannel)
		return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Sender, coins)
	}

	// mint from supply
	err := k.supplyKeeper.MintCoins(ctx, types.GetModuleAccountName(), data.Amount)
	if err != nil {
		return err
	}

	return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GetModuleAccountName(), data.Sender, data.Amount)
}

func (k Keeper) createOutgoingPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string,
	amount sdk.Coins,
	sender sdk.AccAddress,
	receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	if isSourceChain {
		// escrow tokens if the destination chain is the same as the sender's
		escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		coins := make(sdk.Coins, len(amount))
		for i, coin := range amount {
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdkerrors.Wrapf(
					sdkerrors.ErrInvalidCoins,
					"%s doesn't contain the prefix '%s'", coin.Denom, prefix,
				)
			}
			coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
		}

		err := k.bankKeeper.SendCoins(ctx, sender, escrowAddress, coins)
		if err != nil {
			return err
		}

	} else {
		// burn vouchers from the sender's balance if the source is from another chain
		prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		for _, coin := range amount {
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdkerrors.Wrapf(
					sdkerrors.ErrInvalidCoins,
					"%s doesn't contain the prefix '%s'", coin.Denom, prefix,
				)
			}
		}

		// transfer the coins to the module account and burn them
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.GetModuleAccountName(), amount)
		if err != nil {
			return err
		}

		// burn from supply
		err = k.supplyKeeper.BurnCoins(ctx, types.GetModuleAccountName(), amount)
		if err != nil {
			return err
		}
	}

	packetData := types.NewPacketDataTransfer(amount, sender, receiver, isSourceChain, uint64(ctx.BlockHeight())+DefaultPacketTimeout)

	packet := channel.NewPacket(
		packetData,
		seq,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
	)

	return k.channelKeeper.SendPacket(ctx, packet, k.boundedCapability)
}
