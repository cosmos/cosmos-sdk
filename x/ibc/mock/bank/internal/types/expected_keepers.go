package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// IbcBankKeeper expected IBC transfer keeper
type IbcBankKeeper interface {
	ReceiveTransfer(ctx sdk.Context, data types.TransferPacketData, destPort, destChannel, srcPort, srcChannel string) sdk.Error
}

// ChannelKeeper expected IBC channel keeper
type ChannelKeeper interface {
	RecvPacket(
		ctx sdk.Context,
		packet exported.PacketI,
		proof commitment.Proof,
		proofHeight uint64,
		acknowledgement []byte,
	) (exported.PacketI, error)
}
