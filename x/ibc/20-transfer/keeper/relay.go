package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// SendTransfer handles transfer sending logic. There are 2 possible cases:
//
// 1. Sender chain is the source chain of the coins (i.e where they were minted): the coins
// are transferred to an escrow address (i.e locked) on the sender chain and then
// transferred to the destination chain (i.e not the source chain) via a packet
// with the corresponding fungible token data.
//
// 2. Coins are not native from the sender chain (i.e tokens sent where transferred over
// through IBC already): the coins are burned and then a packet is sent to the
// source chain of the tokens.
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	destHeight uint64,
	amount sdk.Coins,
	sender,
	receiver sdk.AccAddress,
	isSourceChain bool, // is the packet sender the source chain of the token?
) error {
	sourceChannelEnd, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return sdkerrors.Wrap(channel.ErrChannelNotFound, sourceChannel)
	}

	destinationPort := sourceChannelEnd.Counterparty.PortID
	destinationChannel := sourceChannelEnd.Counterparty.ChannelID

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channel.ErrSequenceSendNotFound
	}

	return k.createOutgoingPacket(ctx, sequence, sourcePort, sourceChannel, destinationPort, destinationChannel, destHeight, amount, sender, receiver, isSourceChain)
}

// ReceiveTransfer handles transfer receiving logic.
func (k Keeper) ReceiveTransfer(ctx sdk.Context, packet channel.Packet, data types.FungibleTokenPacketData) error {
	return k.onRecvPacket(ctx, packet, data)
}

// TimeoutTransfer handles transfer timeout logic.
func (k Keeper) TimeoutTransfer(ctx sdk.Context, packet channel.Packet, data types.FungibleTokenPacketData) error {
	return k.onTimeoutPacket(ctx, packet, data)
}

// See spec for this function: https://github.com/cosmos/ics/tree/master/spec/ics-020-fungible-token-transfer#packet-relay
func (k Keeper) createOutgoingPacket(
	ctx sdk.Context,
	seq uint64,
	sourcePort, sourceChannel,
	destinationPort, destinationChannel string,
	destHeight uint64,
	amount sdk.Coins,
	sender, receiver sdk.AccAddress,
	isSourceChain bool,
) error {
	// NOTE:
	// - Coins transferred from the destination chain should have their denomination
	// prefixed with source port and channel IDs.
	// - Coins transferred from the source chain can have their denomination
	// clear from prefixes when transferred to the escrow account (i.e when they are
	// locked) BUT MUST have the destination port and channel ID when constructing
	// the packet data.
	var prefix string

	if isSourceChain {
		// clear the denomination from the prefix to send the coins to the escrow account
		coins := make(sdk.Coins, len(amount))
		prefix = types.GetDenomPrefix(destinationPort, destinationChannel)
		for i, coin := range amount {
			if strings.HasPrefix(coin.Denom, prefix) {
				coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
			} else {
				coins[i] = coin
			}
		}

		// escrow tokens if the destination chain is the same as the sender's
		escrowAddress := types.GetEscrowAddress(sourcePort, sourceChannel)

		// escrow source tokens. It fails if balance insufficient.
		if err := k.bankKeeper.SendCoins(
			ctx, sender, escrowAddress, coins,
		); err != nil {
			return err
		}

	} else {
		// build the receiving denomination prefix if it's not present
		prefix = types.GetDenomPrefix(sourcePort, sourceChannel)
		for i, coin := range amount {
			if !strings.HasPrefix(coin.Denom, prefix) {
				amount[i] = sdk.NewCoin(prefix+coin.Denom, coin.Amount)
			}
		}

		// transfer the coins to the module account and burn them
		if err := k.supplyKeeper.SendCoinsFromAccountToModule(
			ctx, sender, types.GetModuleAccountName(), amount,
		); err != nil {
			return err
		}

		// burn vouchers from the sender's balance if the source is from another chain
		if err := k.supplyKeeper.BurnCoins(
			ctx, types.GetModuleAccountName(), amount,
		); err != nil {
			// NOTE: should not happen as the module account was
			// retrieved on the step above and it has enough balace
			// to burn.
			return err
		}
	}

	// NOTE: isSourceChain is negated since the counterparty chain
	//

	packetData := types.NewFungibleTokenPacketData(
		amount, sender, receiver, !isSourceChain, destHeight+DefaultPacketTimeout,
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

func (k Keeper) onRecvPacket(ctx sdk.Context, packet channel.Packet, data types.FungibleTokenPacketData) error {
	// NOTE: packet data type already checked in handler.go

	if data.Source {
		prefix := types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
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

func (k Keeper) onTimeoutPacket(ctx sdk.Context, packet channel.Packet, data types.FungibleTokenPacketData) error {
	// NOTE: packet data type already checked in handler.go

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
