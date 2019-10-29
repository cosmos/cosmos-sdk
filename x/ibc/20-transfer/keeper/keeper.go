package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

const (
	DefaultPacketTimeout = 1000 // default packet timeout relative to the current consensus height of the counterparty
)

// Keeper defines the IBC transfer keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
	prefix    []byte // prefix bytes for accessing the store

	clientKeeper     types.ClientKeeper
	connectionKeeper types.ConnectionKeeper
	channelKeeper    types.ChannelKeeper
	bankKeeper       types.BankKeeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType,
	clientk types.ClientKeeper, connk types.ConnectionKeeper,
	chank types.ChannelKeeper, bk types.BankKeeper,
) Keeper {
	return Keeper{
		storeKey:         key,
		cdc:              cdc,
		codespace:        sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/transfer",
		prefix:           []byte(types.SubModuleName + "/"),                                          // "transfer/"
		clientKeeper:     clientk,
		connectionKeeper: connk,
		channelKeeper:    chank,
		bankKeeper:       bk,
	}
}

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
	//TODO
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
		_, err := k.bankKeeper.AddCoins(ctx, data.Sender, data.Amount)
		if err != nil {
			return err
		}
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

// SendTransfer handles transfer sending logic
func (k Keeper) SendTransfer(
	ctx sdk.Context,
	sourcePort,
	sourceChannel string,
	amount sdk.Coins,
	sender,
	receiver sdk.AccAddress,
	isSourceChain bool,
) sdk.Error {
	// get the port and channel of the counterparty
	channel, found := k.channelKeeper.GetChannel(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrChannelNotFound(k.codespace, sourceChannel)
	}

	destinationPort := channel.Counterparty.PortID
	destinationChannel := channel.Counterparty.ChannelID

	// get the next sequence
	sequence, found := k.channelKeeper.GetNextSequenceSend(ctx, sourcePort, sourceChannel)
	if !found {
		return channeltypes.ErrSequenceNotFound(k.codespace, "send")
	}

	coins := make(sdk.Coins, len(amount))
	prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
	switch {
	case isSourceChain:
		// build the receiving denomination prefix
		for i, coin := range amount {
			coin := coin
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
	data types.TransferPacketData,
) sdk.Error {
	if data.Source {
		// mint tokens

		// check the denom prefix
		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		for _, coin := range data.Amount {
			coin := coin
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
			}
		}

		// TODO: use supply keeper to mint
		_, err := k.bankKeeper.AddCoins(ctx, data.Receiver, data.Amount)
		return err
	}

	// unescrow tokens

	// check the denom prefix
	prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
	coins := make(sdk.Coins, len(data.Amount))
	for i, coin := range data.Amount {
		coin := coin
		if !strings.HasPrefix(coin.Denom, prefix) {
			return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
		}
		coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
	}

	escrowAddress := types.GetEscrowAddress(destinationChannel)
	return k.bankKeeper.SendCoins(ctx, escrowAddress, data.Receiver, coins)

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
) sdk.Error {
	if isSourceChain {
		// escrow tokens

		// get escrow address
		escrowAddress := types.GetEscrowAddress(sourceChannel)

		// check the denom prefix
		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		coins := make(sdk.Coins, len(amount))
		for i, coin := range amount {
			coin := coin
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
			}
			coins[i] = sdk.NewCoin(coin.Denom[len(prefix):], coin.Amount)
		}

		err := k.bankKeeper.SendCoins(ctx, sender, escrowAddress, coins)
		if err != nil {
			return err
		}

	} else {
		// burn vouchers from sender

		// check the denom prefix
		prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		for _, coin := range amount {
			coin := coin
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
			}
		}

		// TODO: use supply keeper to burn
		_, err := k.bankKeeper.SubtractCoins(ctx, sender, amount)
		if err != nil {
			return err
		}
	}

	packetData := types.TransferPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   isSourceChain,
	}

	packetDataBz, err := packetData.MarshalJSON()
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidPacketData, "invalid packet data")
	}

	packet := channeltypes.NewPacket(
		seq,
		uint64(ctx.BlockHeight())+DefaultPacketTimeout,
		sourcePort,
		sourceChannel,
		destinationPort,
		destinationChannel,
		packetDataBz,
	)

	err = k.channelKeeper.SendPacket(ctx, packet, k.storeKey)
	if err != nil {
		return types.ErrSendPacket(types.DefaultCodespace)
	}

	return nil
}
