package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/17-nft_transfer/types"
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

	err := data.UnmarshalJSON(packet.Data())
	if err != nil {
		return types.ErrInvalidPacketData(k.codespace)
	}

	return k.ReceiveTransfer(
		ctx, packet.SourcePort(), packet.SourceChannel(),
		packet.DestPort(), packet.DestChannel(), data,
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

	err := data.UnmarshalJSON(packet.Data())
	if err != nil {
		return types.ErrInvalidPacketData(k.codespace)
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.SourcePort(), packet.SourcePort())

	if !strings.HasPrefix(data.Denom, prefix) {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidDenom, "incorrect denomination")
	}
	id := data.ID

	if data.Source {
		escrowAddress := types.GetEscrowAddress(packet.DestPort(), packet.DestChannel())
		denom := data.Denom[len(prefix):]
		nft, err := k.nftKeeper.GetNFT(ctx, denom, id)
		if err != nil {
			return err
		}
		// NFT needs to be in escrow to continue
		if !nft.GetOwner().Equals(escrowAddress) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidNFT, "NFT should be in escrow")
		}
		// update NFT owner
		nft.SetOwner(data.Sender)
		return k.nftKeeper.UpdateNFT(ctx, denom, nft)
	}
	// mint from supply
	nft := NewBaseNFT(data.ID, data.Sender, data.TokenURI)
	return k.nftKeeper.MintNFT(ctx, data.Denom, &nft)
}

// nolint: unused
func (k Keeper) onTimeoutPacketClose(_ sdk.Context, _ channeltypes.Packet) {
	panic("can't happen, only unordered channels allowed")
}
