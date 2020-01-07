package keeper

// NOTE:
// OnChanOpenInit, OnChanOpenTry, OnChanOpenAck, OnChanOpenConfirm, OnChanCLoseConfirm
// will be implemented according to ADR15 in the future PRs. Code left for reference.
//
// OnRecvPacket, OnAcknowledgementPacket, OnTimeoutPacket has been implemented according
// to ADR15.

/*
import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// nolint: unused
func (k Keeper) OnChanOpenInit(
	ctx sdk.Context,
	order channel.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channel.Counterparty,
	version string,
) error {
	if order != channel.UNORDERED {
		return sdkerrors.Wrap(channel.ErrInvalidChannel, "channel must be UNORDERED")
	}

	// NOTE: here the capability key name defines the port ID of the counterparty
	if counterparty.PortID != k.boundedCapability.Name() {
		return sdkerrors.Wrapf(
			port.ErrInvalidPort,
			"counterparty port ID doesn't match the capability key (%s ≠ %s)", counterparty.PortID, k.boundedCapability.Name(),
		)
	}

	if strings.TrimSpace(version) != "" {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "version must be blank")
	}

	// NOTE: as the escrow address is generated from both the port and channel IDs
	// there's no need to store it on a map.
	return nil
}

// nolint: unused
func (k Keeper) OnChanOpenTry(
	ctx sdk.Context,
	order channel.Order,
	connectionHops []string,
	portID,
	channelID string,
	counterparty channel.Counterparty,
	version string,
	counterpartyVersion string,
) error {
	if order != channel.UNORDERED {
		return sdkerrors.Wrap(channel.ErrInvalidChannel, "channel must be UNORDERED")
	}

	// NOTE: here the capability key name defines the port ID of the counterparty
	if counterparty.PortID != k.boundedCapability.Name() {
		return sdkerrors.Wrapf(
			port.ErrInvalidPort,
			"counterparty port ID doesn't match the capability key (%s ≠ %s)", counterparty.PortID, k.boundedCapability.Name(),
		)
	}

	if strings.TrimSpace(version) != "" {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "version must be blank")
	}

	if strings.TrimSpace(counterpartyVersion) != "" {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "counterparty version must be blank")
	}

	// NOTE: as the escrow address is generated from both the port and channel IDs
	// there's no need to store it on a map.
	return nil
}

// nolint: unused
func (k Keeper) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	version string,
) error {
	if strings.TrimSpace(version) != "" {
		return sdkerrors.Wrap(ibctypes.ErrInvalidVersion, "version must be blank")
	}

	return nil
}

// nolint: unused
func (k Keeper) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// no-op
	return nil
}

// nolint: unused
func (k Keeper) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// no-op
	return nil
}

// nolint: unused
func (k Keeper) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	// no-op
	return nil
}

// nolint: unused
func (k Keeper) OnTimeoutPacket(
	ctx sdk.Context,
	packet channelexported.PacketI,
) error {
	var data types.PacketData

	err := k.cdc.UnmarshalBinaryBare(packet.GetData(), &data)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid packet data")
	}

	// check the denom prefix
	prefix := types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		coin := coin
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdkerrors.Wrapf(
				sdkerrors.ErrInvalidCoins,
				"%s doesn't contain the prefix '%s'", coin.Denom, prefix,
			)
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
*/
