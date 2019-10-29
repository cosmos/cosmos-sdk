package ibc

// // TODO: Should extract timeout msgs too
// func ExtractMsgPackets(msgs []sdk.Msg) (res []MsgPacket, abort bool) {
// 	res = make([]MsgPacket, 0, len(msgs))
// 	for _, msg := range msgs {
// 		msgp, ok := msg.(MsgPacket)
// 		if ok {
// 			res = append(res, msgp)
// 		}
// 	}

// 	if len(res) >= 2 {
// 		first := res[0]
// 		for _, msg := range res[1:] {
// 			if len(msg.ChannelID) != 0 && msg.ChannelID != first.ChannelID {
// 				return res, true
// 			}
// 			msg.ChannelID = first.ChannelID
// 		}
// 	}

// 	return
// }

// func VerifyMsgPackets(ctx sdk.Context, channel channel.Manager, msgs []MsgPacket) error {
// 	for _, msg := range msgs {
// 		err := channel.Receive(ctx, msg.Proofs, msg.Height, msg.ReceiverPort(), msg.ChannelID, msg.Packet)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func NewAnteDecorator(channel channel.Manager) sdk.AnteDecorator {
// 	return func(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
// 		msgs, abort := ExtractMsgPackets(tx.GetMsgs())
// 		if abort {
// 			return ctx, host.ErrInvalidPacket
// 		}

// 		err := VerifyMsgPackets(ctx, channel, msgs)
// 		if err != nil {
// 			return ctx, sdkerrors.Wrap(host.ErrInvalidPacket, err.Error())
// 		}

// 		return next(ctx, tx, simulate)
// 	}
// }
