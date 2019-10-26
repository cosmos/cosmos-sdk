package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	"github.com/tendermint/tendermint/crypto"
)

const (
	DefaultPacketTimeout = 1000 // default packet timeout relative to the current block height
)

// Keeper defines the IBC transfer keeper
type Keeper struct {
	cdc *codec.Codec
	key sdk.StoreKey
	ck  types.ChannelKeeper
	bk  types.BankKeeper
}

// NewKeeper creates a new IBC transfer Keeper instance
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ck types.ChannelKeeper, bk types.BankKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
		ck:  ck,
		bk:  bk,
	}
}

// SendTransfer handles transfer sending logic
func (k Keeper) SendTransfer(ctx sdk.Context, srcPort, srcChan string, denom string, amount sdk.Int, sender sdk.AccAddress, receiver string, source bool) sdk.Error {
	// get the port and channel of the counterparty
	channel, ok := k.ck.GetChannel(ctx, srcPort, srcChan)
	if !ok {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), channeltypes.CodeChannelNotFound, "failed to get channel")
	}

	destPort := channel.Counterparty.PortID
	destChan := channel.Counterparty.ChannelID

	// get the next sequence
	sequence, ok := k.ck.GetNextSequenceSend(ctx, srcPort, srcChan)
	if !ok {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), channeltypes.CodeSequenceNotFound, "failed to retrieve sequence")
	}

	if source {
		// build the receiving denomination prefix
		denom = getDenomPrefix(destPort, destChan) + denom
	}

	return k.createOutgoingPacket(ctx, sequence, srcPort, srcChan, destPort, destChan, denom, amount, sender, receiver, source)
}

// ReceiveTransfer handles transfer receiving logic
func (k Keeper) ReceiveTransfer(ctx sdk.Context, data types.TransferPacketData, destPort, destChannel, srcPort, srcChannel string) sdk.Error {

	receiverAddr, err := sdk.AccAddressFromBech32(data.Receiver)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidReceiver, "invalid receiver address")
	}

	if data.Source {
		// mint tokens

		// check the denom prefix
		if !strings.HasPrefix(data.Denomination, getDenomPrefix(destPort, destChannel)) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeIncorrectDenom, "incorrect denomination")
		}

		_, err := k.bk.AddCoins(ctx, receiverAddr, sdk.Coins{sdk.NewCoin(data.Denomination, data.Amount)})
		if err != nil {
			return err
		}

	} else {
		// unescrow tokens

		// check the denom prefix
		prefix := getDenomPrefix(srcPort, srcChannel)
		if !strings.HasPrefix(data.Denomination, prefix) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeIncorrectDenom, "incorrect denomination")
		}

		escrowAddress := getEscrowAddress(destChannel)
		err := k.bk.SendCoins(ctx, escrowAddress, receiverAddr, sdk.Coins{sdk.NewCoin(data.Denomination[len(prefix):], data.Amount)})
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) createOutgoingPacket(ctx sdk.Context, seq uint64, srcPort, srcChan, destPort, destChan string, denom string, amount sdk.Int, sender sdk.AccAddress, receiver string, source bool) sdk.Error {
	if source {
		// escrow tokens

		// get escrow address
		escrowAddress := getEscrowAddress(srcChan)

		// check the denom prefix
		prefix := getDenomPrefix(destPort, destChan)
		if !strings.HasPrefix(denom, prefix) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeIncorrectDenom, "incorrect denomination")
		}

		err := k.bk.SendCoins(ctx, sender, escrowAddress, sdk.Coins{sdk.NewCoin(denom[len(prefix):], amount)})
		if err != nil {
			return err
		}

	} else {
		// burn vouchers from sender

		// check the denom prefix
		prefix := getDenomPrefix(srcPort, srcChan)
		if !strings.HasPrefix(denom, prefix) {
			return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeIncorrectDenom, "incorrect denomination")
		}

		_, err := k.bk.SubtractCoins(ctx, sender, sdk.Coins{sdk.NewCoin(denom, amount)})
		if err != nil {
			return err
		}
	}

	// build packet
	packetData := types.TransferPacketData{
		Denomination: denom,
		Amount:       amount,
		Sender:       sender.String(),
		Receiver:     receiver,
		Source:       source,
	}

	packetDataBz, err := packetData.MarshalJSON()
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidPacketData, "invalid packet data")
	}

	packet := channeltypes.NewPacket(seq, uint64(ctx.BlockHeight())+DefaultPacketTimeout, srcPort, srcChan, destPort, destChan, packetDataBz)

	err = k.ck.SendPacket(ctx, packet)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeErrSendPacket, "failed to send packet")
	}

	return nil
}

// getEscrowAddress returns the escrow address for the specified channel
func getEscrowAddress(chanID string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(chanID)))
}

// getDenomPrefix returns the receiving denomination prefix
func getDenomPrefix(port, channel string) string {
	return fmt.Sprintf("%s/%s", port, channel)
}
