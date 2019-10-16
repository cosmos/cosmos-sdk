package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	chantypes "github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
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

// Transfer handles transfer logic
func (k Keeper) Transfer(ctx sdk.Context, srcPort, srcChan, dstPort, dstChan string, amount sdk.Coin, sender, receiver sdk.AccAddress, source bool) sdk.Error {
	// TODO
	if source {
		// escrow tokens
		escrowAddress := k.GetEscrowAddress(srcChan)
		k.bk.SendCoins(ctx, sender, escrowAddress, sdk.Coins{amount})
	} else {
		// burn vouchers from sender
		k.bk.BurnCoins(ctx, sender, sdk.Coins{amount})
	}

	// get the next sequence
	sequence, ok := k.ibck.ChannelKeeper.GetNextSequenceSend(ctx, srcPort, srcChan)
	if !ok {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeErrGetSequence, "failed to retrieve sequence")
	}

	// build packet
	packetData := types.TransferPacketData{
		Amount:   amount,
		Sender:   sender,
		Receiver: receiver,
		Source:   source,
	}

	packet := chantypes.NewPacket(sequence, 0, srcPort, srcChan, dstPort, dstChan, packetData.Marshal())

	err := k.ibck.ChannelKeeper.SendPacket(ctx, packet)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeErrSendPacket, "failed to send packet")
	}

	return nil
}

// GetEscrowAddress returns the escrow address for the specified channel
func (k Keeper) GetEscrowAddress(chanID string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(chanID)))
}
