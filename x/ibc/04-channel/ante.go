package channel

import (
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
	for _, msg := range tx.GetMsgs() {
		var err error
		switch msg := msg.(type) {
		case client.MsgUpdateClient:
			err = pvr.clientKeeper.UpdateClient(ctx, msg.ClientID, msg.Header)
		case MsgPacket:
			_, err = pvr.channelKeeper.RecvPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight)
		case MsgAcknowledgement:
			_, err = pvr.channelKeeper.AcknowledgePacket(ctx, msg.Packet, msg.Acknowledgement, msg.Proof, msg.ProofHeight)
		case MsgTimeout:
			_, err = pvr.channelKeeper.TimeoutPacket(ctx, msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
		}

		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
