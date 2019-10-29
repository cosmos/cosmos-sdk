package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

func (k Keeper) onChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
) error {
	if order != channeltypes.UNORDERED {
		return types.ErrInvalidChannelOrder(types.DefaultCodespace, order.String())
	}

	if counterparty.PortID != types.BoundPortID {
		return types.ErrInvalidPort(types.DefaultCodespace, portID)
	}

	if version != "" {
		return types.ErrInvalidVersion(types.DefaultCodespace, fmt.Sprintf("invalid version: %s", version))
	}

	return nil
}

func (k Keeper) onChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channeltypes.Counterparty,
	version string,
	counterpartyVersion string,
) error {
	if order != channeltypes.UNORDERED {
		return types.ErrInvalidChannelOrder(types.DefaultCodespace, order.String())
	}

	if counterparty.PortID != types.BoundPortID {
		return types.ErrInvalidPort(types.DefaultCodespace, portID)
	}

	if version != "" {
		return types.ErrInvalidVersion(types.DefaultCodespace, fmt.Sprintf("invalid version: %s", version))
	}

	if counterpartyVersion != "" {
		return types.ErrInvalidVersion(types.DefaultCodespace, fmt.Sprintf("invalid counterparty version: %s", version))
	}

	return nil
}

func (k Keeper) onChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	version string,
) error {
	if version != "" {
		return types.ErrInvalidVersion(types.DefaultCodespace, fmt.Sprintf("invalid version: %s", version))
	}

	return nil
}

func (k Keeper) onChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// noop
	return nil
}

func (k Keeper) onChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// noop
	return nil
}

func (k Keeper) onChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// noop
	return nil
}

// onRecvPacket is called when an FTTransfer packet is received
func (k Keeper) onRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	var data types.TransferPacketData

	err := data.UnmarshalJSON(packet.Data())
	if err != nil {
		return types.ErrInvalidPacketData(types.DefaultCodespace)
	}

	return k.ReceiveTransfer(ctx, packet.SourcePort(), packet.SourceChannel(),
		packet.DestPort(), packet.DestChannel(), data)
}

func (k Keeper) onAcknowledgePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
) error {
	// noop
	return nil
}

func (k Keeper) onTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	var data types.TransferPacketData

	err := data.UnmarshalJSON(packet.Data())
	if err != nil {
		return types.ErrInvalidPacketData(types.DefaultCodespace)
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.SourcePort(), packet.SourcePort())
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		coin := coin
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
		}
		coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	}

	if data.Source {
		escrowAddress := types.GetEscrowAddress(packet.DestChannel())

		err := k.bankKeeper.SendCoins(ctx, escrowAddress, data.Sender, coins)
		if err != nil {
			return err
		}

	} else {
		// mint from supply
		err = k.supplyKeeper.MintCoins(ctx, types.GetModuleAccountName(), data.Amount)
		if err != nil {
			return err
		}

		return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GetModuleAccountName(), data.Sender, data.Amount)
	}

	return nil
}

func (k Keeper) onTimeoutPacketClose(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	// noop
	return nil
}
