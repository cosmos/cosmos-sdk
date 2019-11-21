package ibc

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
)

type ProofVerificationDecorator struct {
	clientKeeper  ClientKeeper
	channelKeeper ChannelKeeper
}

func (pvr ProofVerificationDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	var flag bool

	for _, msg := range tx.GetMsgs() {
		var err error
		switch msg := msg.(type) {
		case client.MsgUpdateClient:
			err = pvr.clientKeeper.UpdateClient(msg.ClientID, msg.Header)
			flag = true
		case channel.MsgPacket:
			err = pvr.channelKeeper.VerifyPacket(msg.Packet, msg.Proofs, msg.ProofHeight)
			flag = true
			pvr.channelKeeper.SetPacketAcknowledgement(ctx, msg.PortID, msg.ChannelID, msg.Sequence, []byte{})
		case channel.MsgAcknowledgement:
			err = pvr.channelKeeper.VerifyAcknowledgement(msg.Acknowledgement, msg.Proof, msg.ProofHeight)
			flag = true
		case channel.MsgTimeoutPacket:
			err = pvr.channelKeeper.VerifyTimeout(msg.Packet, msg.Proof, msg.ProofHeight, msg.NextSequenceRecv)
			flag = true
		default:
			err = errors.New("Transaction cannot include both IBC packet messages and normal messages")
		}

		if err != nil {
			return ctx, err
		}
	}

	return next(ctx, tx, simulate)
}
