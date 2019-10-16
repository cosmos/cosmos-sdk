package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	chantypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/types"
	"github.com/tendermint/tendermint/crypto"
)

type Keeper struct {
	cdc  *codec.Codec
	key  sdk.StoreKey
	ibck ibc.Keeper
	bk   types.BankKeeper
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ibck ibc.Keeper, bk types.BankKeeper) Keeper {
	return Keeper{
		cdc:  cdc,
		key:  key,
		ibck: ibck,
		bk:   bk,
	}
}

// SendTransfer handles transfer sending logic
func (k Keeper) SendTransfer(ctx sdk.Context, srcPort, srcChan string, amount sdk.Coin, sender sdk.AccAddress, receiver string, source bool, timeout uint64) sdk.Error {
	// get the port and channel of the counterparty
	channel, ok := k.ibck.ChannelKeeper.GetChannel(ctx, srcPort, srcChan)
	if !ok {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), chantypes.CodeChannelNotFound, "failed to get channel")
	}

	dstPort := channel.Counterparty.PortID
	dstChan := channel.Counterparty.ChannelID

	// get the next sequence
	sequence, ok := k.ibck.ChannelKeeper.GetNextSequenceSend(ctx, srcPort, srcChan)
	if !ok {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), chantypes.CodeSequenceNotFound, "failed to retrieve sequence")
	}

	if source {
		// escrow tokens
		escrowAddress := k.GetEscrowAddress(srcChan)
		err := k.bk.SendCoins(ctx, sender, escrowAddress, sdk.Coins{amount})
		if err != nil {
			return err
		}

	} else {
		// burn vouchers from sender
		err := k.bk.BurnCoins(ctx, sender, sdk.Coins{amount})
		if err != nil {
			return err
		}
	}

	// build packet
	packetData := types.TransferPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
	}

	packet := chantypes.NewPacket(sequence, timeout, srcPort, srcChan, dstPort, dstChan, packetData.Marshal())

	err := k.ibck.ChannelKeeper.SendPacket(ctx, packet)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeErrSendPacket, "failed to send packet")
	}

	return nil
}

// ReceiveTransfer handles transfer receiving logic
func (k Keeper) ReceiveTransfer(ctx sdk.Context, srcPort, srcChan string, amount sdk.Coin, sender sdk.AccAddress, receiver string, source bool, timeout uint64, proof ics23.Proof, proofHeight uint64) sdk.Error {
	// get the port and channel of the counterparty
	channel, ok := k.ibck.ChannelKeeper.GetChannel(ctx, srcPort, srcChan)
	if !ok {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), chantypes.CodeChannelNotFound, "failed to get channel")
	}

	dstPort := channel.Counterparty.PortID
	dstChan := channel.Counterparty.ChannelID

	receiverAddr, err := sdk.AccAddressFromBech32(receiver)
	if err != nil {
		sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidReceiver, "invalid receiver address")
	}

	if source {
		// mint tokens
		_, err := k.bk.AddCoins(ctx, receiverAddr, sdk.Coins{amount})
		if err != nil {
			return err
		}

	} else {
		// unescrow tokens
		escrowAddress := k.GetEscrowAddress(dstChan)
		err := k.bk.SendCoins(ctx, escrowAddress, receiverAddr, sdk.Coins{amount})
		if err != nil {
			return err
		}
	}

	// build packet
	packetData := types.TransferPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
	}

	sequence := uint64(0) // unordered channel
	packet := chantypes.NewPacket(sequence, timeout, srcPort, srcChan, dstPort, dstChan, packetData.Marshal())

	_, err = k.ibck.ChannelKeeper.RecvPacket(ctx, packet, proof, proofHeight, nil)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeErrReceivePacket, "failed to receive packet")
	}

	return nil
}

// GetEscrowAddress returns the escrow address for the specified channel
func (k Keeper) GetEscrowAddress(chanID string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(chanID)))
}
