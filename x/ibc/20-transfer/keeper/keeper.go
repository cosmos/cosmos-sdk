package keeper

import (
	"fmt"
	"strings"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// DefaultPacketTimeout is the default packet timeout relative to the current block height
const (
	DefaultPacketTimeout = 1000 // NOTE: in blocks
)

// Keeper defines the IBC transfer keeper
type Keeper struct {
	storeKey  sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType

	bankKeeper    types.BankKeeper
	channelKeeper types.ChannelKeeper
	supplyKeeper  types.SupplyKeeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(
	cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType,
	bk types.BankKeeper, ck types.ChannelKeeper, sk types.SupplyKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      key,
		codespace:     sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/transfer"
		bankKeeper:    bk,
		channelKeeper: ck,
		supplyKeeper:  sk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// GetTransferAccount returns the ICS20 - transfers ModuleAccount
func (k Keeper) GetTransferAccount(ctx sdk.Context) supplyexported.ModuleAccountI {
	return k.supplyKeeper.GetModuleAccount(ctx, types.ModuleAccountName)
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
	data types.TransferPacketData,
	sourcePort,
	sourceChannel,
	destinationPort,
	destinationChannel string) sdk.Error {

	if data.Source {
		// mint new tokens if the source of the transfer is the same chain
		err := k.supplyKeeper.MintCoins(ctx, types.ModuleAccountName, data.Amount)
		if err != nil {
			return err
		}

		prefix := types.GetDenomPrefix(destinationPort, destinationChannel)
		for _, coin := range data.Amount {
			coin := coin
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
			}
		}

		return k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleAccountName, data.Receiver, data.Amount)
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
	isSourceChain bool) sdk.Error {
	if isSourceChain {
		// escrow tokens if the destination chain is the same as the sender's
		escrowAddress := types.GetEscrowAddress(sourceChannel)

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
		// burn vouchers from the sender's balance if the source is from another chain
		prefix := types.GetDenomPrefix(sourcePort, sourceChannel)
		for _, coin := range amount {
			coin := coin
			if !strings.HasPrefix(coin.Denom, prefix) {
				return sdk.ErrInvalidCoins(fmt.Sprintf("%s doesn't contain the prefix '%s'", coin.Denom, prefix))
			}
		}

		// transfer the coins to the module account and burn them
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, sender, types.ModuleAccountName, amount)
		if err != nil {
			return err
		}

		err = k.supplyKeeper.BurnCoins(ctx, types.ModuleAccountName, amount)
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

	return k.channelKeeper.SendPacket(ctx, packet, k.storeKey)
}
