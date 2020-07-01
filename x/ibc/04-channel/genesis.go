package channel

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/keeper"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
)

// InitGenesis initializes the ibc channel submodule's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, gs types.GenesisState) {
	for _, channel := range gs.Channels {
		ch := types.NewChannel(channel.State, channel.Ordering, channel.Counterparty, channel.ConnectionHops, channel.Version)
		k.SetChannel(ctx, channel.PortID, channel.ID, ch)
	}
	for _, ack := range gs.Acknowledgements {
		k.SetPacketAcknowledgement(ctx, ack.PortID, ack.ChannelID, ack.Sequence, ack.Hash)
	}
	for _, commitment := range gs.Commitments {
		k.SetPacketCommitment(ctx, commitment.PortID, commitment.ChannelID, commitment.Sequence, commitment.Hash)
	}
	for _, ss := range gs.SendSequences {
		k.SetNextSequenceSend(ctx, ss.PortID, ss.ChannelID, ss.Sequence)
	}
	for _, rs := range gs.RecvSequences {
		k.SetNextSequenceRecv(ctx, rs.PortID, rs.ChannelID, rs.Sequence)
	}
	for _, as := range gs.AckSequences {
		k.SetNextSequenceAck(ctx, as.PortID, as.ChannelID, as.Sequence)
	}
}

// ExportGenesis returns the ibc channel submodule's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	return types.GenesisState{
		Channels:         k.GetAllChannels(ctx),
		Acknowledgements: k.GetAllPacketAcks(ctx),
		Commitments:      k.GetAllPacketCommitments(ctx),
		SendSequences:    k.GetAllPacketSendSeqs(ctx),
		RecvSequences:    k.GetAllPacketRecvSeqs(ctx),
		AckSequences:     k.GetAllPacketAckSeqs(ctx),
	}
}
