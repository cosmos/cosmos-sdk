package channel

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

type ProofVerificationDecorator struct {
	clientKeeper  client.Keeper
	channelKeeper Keeper
}

func NewProofVerificationDecorator(clientKeeper client.Keeper, channelKeeper Keeper) ProofVerificationDecorator {
	return ProofVerificationDecorator{
		clientKeeper:  clientKeeper,
		channelKeeper: channelKeeper,
	}
}

func (pvr ProofVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	var flag bool
	var portID, channelID string

	for _, msg := range tx.GetMsgs() {
		var err error
		switch msg := msg.(type) {
		case client.MsgUpdateClient:
			err = pvr.clientKeeper.UpdateClient(ctx, msg.ClientID, msg.Header)
		case MsgPacket:
			err = pvr.channelKeeper.RecvPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight)
			if flag {
				if portID != msg.DestinationPort || channelID != msg.DestinationChannel {
					return ctx, errors.New("Transaction cannot include IBC packets from different channels")
				}
			} else {
				portID = msg.DestinationPort
				channelID = msg.DestinationChannel
				flag = true
			}
			pvr.channelKeeper.SetPacketAcknowledgement(ctx, msg.DestinationPort, msg.DestinationChannel, msg.Sequence, []byte{})

		case MsgAcknowledgement:
			err = pvr.channelKeeper.AcknowledgementPacket(ctx, msg.Packet, msg.Acknowledgement, msg.Proof, msg.ProofHeight)
			if flag {
				if portID != msg.SourcePort || channelID != msg.SourceChannel {
					return ctx, errors.New("Transaction cannot include IBC packets from different channels")
				}
			} else {
				portID = msg.SourcePort
				channelID = msg.SourceChannel
				flag = true
			}
		case MsgTimeout:
			err = pvr.channelKeeper.TimeoutPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
			if flag {
				if portID != msg.SourcePort || channelID != msg.SourceChannel {
					return ctx, errors.New("Transaction cannot include IBC packets from different channels")
				}
			} else {
				portID = msg.SourcePort
				channelID = msg.SourceChannel
				flag = true
			}
		default:
			err = errors.New("Transaction cannot include both IBC packet messages and normal messages")
		}

		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
