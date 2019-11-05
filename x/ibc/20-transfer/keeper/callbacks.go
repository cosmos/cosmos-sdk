package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// nolint: unused
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
		return types.ErrInvalidChannelOrder(k.codespace, order.String())
	}

	if counterparty.PortID != types.BoundPortID {
		return types.ErrInvalidPort(k.codespace, portID)
	}

	if strings.TrimSpace(version) != "" {
		return types.ErrInvalidVersion(k.codespace, fmt.Sprintf("invalid version: %s", version))
	}

	// NOTE: as the escrow address is generated from both the port and channel IDs
	// there's no need to store it on a map.
	return nil
}

// nolint: unused
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
		return types.ErrInvalidChannelOrder(k.codespace, order.String())
	}

	if counterparty.PortID != types.BoundPortID {
		return types.ErrInvalidPort(k.codespace, portID)
	}

	if strings.TrimSpace(version) != "" {
		return types.ErrInvalidVersion(k.codespace, fmt.Sprintf("invalid version: %s", version))
	}

	if strings.TrimSpace(counterpartyVersion) != "" {
		return types.ErrInvalidVersion(k.codespace, fmt.Sprintf("invalid counterparty version: %s", version))
	}

	// NOTE: as the escrow address is generated from both the port and channel IDs
	// there's no need to store it on a map.
	return nil
}

// nolint: unused
func (k Keeper) onChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	version string,
) error {
	if strings.TrimSpace(version) != "" {
		return types.ErrInvalidVersion(k.codespace, fmt.Sprintf("invalid version: %s", version))
	}

	return nil
}

// nolint: unused
func (k Keeper) onChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// no-op
	return nil
}

// nolint: unused
func (k Keeper) onChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// no-op
	return nil
}

// nolint: unused
func (k Keeper) onChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// no-op
	return nil
}

// onRecvPacket is called when an FTTransfer packet is received
// nolint: unused
func (k Keeper) onRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	var data types.PacketData

	err := data.UnmarshalJSON(packet.GetData())
	if err != nil {
		return types.ErrInvalidPacketData(k.codespace)
	}

	return k.ReceiveTransfer(
		ctx, packet.GetSourcePort(), packet.GetSourceChannel(),
		packet.GetDestPort(), packet.GetDestChannel(), data,
	)
}

// nolint: unused
func (k Keeper) onAcknowledgePacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
) error {
	// no-op
	return nil
}

// nolint: unused
func (k Keeper) onTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	var data types.PacketData

	err := data.UnmarshalJSON(packet.GetData())
	if err != nil {
		return types.ErrInvalidPacketData(k.codespace)
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourcePort())
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		coin := coin
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
		}
		coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	}

	if data.Source {
		escrowAddress := types.GetEscrowAddress(packet.GetDestPort(), packet.GetDestChannel())
		return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Sender, coins)
	}

	// mint from supply
	err = k.supplyKeeper.MintCoins(ctx, types.GetModuleAccountName(), data.Amount)
	if err != nil {
		return err
	}

	return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.GetModuleAccountName(), data.Sender, data.Amount)
}

// nolint: unused
func (k Keeper) onTimeoutPacketClose(_ sdk.Context, _ channeltypes.Packet) {
	panic("can't happen, only unordered channels allowed")
}
