package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	channelkeeper "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/keeper"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

type Decorator struct {
	k channelkeeper.Keeper
}

func NewAnteDecorator(k channelkeeper.Keeper) Decorator {
	return Decorator{k: k}
}

// AnteDecorator returns an error if a multiMsg tx only contains packet messages (Recv, Ack, Timeout) and additional update messages and all packet messages
// are redundant. If the transaction is just a single UpdateClient message, or the multimsg transaction contains some other message type, then the antedecorator returns no error
// and continues processing to ensure these transactions are included.
// This will ensure that relayers do not waste fees on multiMsg transactions when another relayer has already submitted all packets, by rejecting the tx at the mempool layer.
func (ad Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	// do not run redundancy check on DeliverTx or simulate
	if (ctx.IsCheckTx() || ctx.IsReCheckTx()) && !simulate {
		// keep track of total packet messages and number of redundancies across `RecvPacket`, `AcknowledgePacket`, and `TimeoutPacket/OnClose`
		redundancies := 0
		packetMsgs := 0
		for _, m := range tx.GetMsgs() {
			switch msg := m.(type) {
			case *channeltypes.MsgRecvPacket:
				if _, found := ad.k.GetPacketReceipt(ctx, msg.Packet.GetDestPort(), msg.Packet.GetDestChannel(), msg.Packet.GetSequence()); found {
					redundancies++
				}
				packetMsgs++

			case *channeltypes.MsgAcknowledgement:
				if commitment := ad.k.GetPacketCommitment(ctx, msg.Packet.GetSourcePort(), msg.Packet.GetSourceChannel(), msg.Packet.GetSequence()); len(commitment) == 0 {
					redundancies++
				}
				packetMsgs++

			case *channeltypes.MsgTimeout:
				if commitment := ad.k.GetPacketCommitment(ctx, msg.Packet.GetSourcePort(), msg.Packet.GetSourceChannel(), msg.Packet.GetSequence()); len(commitment) == 0 {
					redundancies++
				}
				packetMsgs++

			case *channeltypes.MsgTimeoutOnClose:
				if commitment := ad.k.GetPacketCommitment(ctx, msg.Packet.GetSourcePort(), msg.Packet.GetSourceChannel(), msg.Packet.GetSequence()); len(commitment) == 0 {
					redundancies++
				}
				packetMsgs++

			case *clienttypes.MsgUpdateClient:
				// do nothing here, as we want to avoid updating clients if it is batched with only redundant messages

			default:
				// if the multiMsg tx has a msg that is not a packet msg or update msg, then we will not return error
				// regardless of if all packet messages are redundant. This ensures that non-packet messages get processed
				// even if they get batched with redundant packet messages.
				return next(ctx, tx, simulate)
			}

		}

		// only return error if all packet messages are redundant
		if redundancies == packetMsgs && packetMsgs > 0 {
			return ctx, channeltypes.ErrRedundantTx
		}
	}
	return next(ctx, tx, simulate)
}
