package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	ics20 "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/mock/bank/internal/types"
)

type Keeper struct {
	cdc *codec.Codec
	key sdk.StoreKey
	ck  types.ChannelKeeper
	bk  types.IbcBankKeeper
}

func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, ck types.ChannelKeeper, bk types.IbcBankKeeper) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
		ck:  ck,
		bk:  bk,
	}
}

// ReceiveTransferPacket handles transfer receiving packet
func (k Keeper) ReceiveTransferPacket(ctx sdk.Context, packet exported.PacketI, proof ics23.Proof, height uint64) sdk.Error {
	_, err := k.ck.RecvPacket(ctx, packet, proof, height, nil)
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeErrReceivePacket, "failed to receive packet")
	}

	var data ics20.TransferPacketData
	err = data.UnmarshalJSON(packet.Data())
	if err != nil {
		return sdk.NewError(sdk.CodespaceType(types.DefaultCodespace), types.CodeInvalidPacketData, "invalid packet data")
	}

	return k.bk.ReceiveTransfer(ctx, data, packet.DestPort(), packet.DestChannel(), packet.SourcePort(), packet.DestChannel())
}
