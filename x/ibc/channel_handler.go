package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type SendHandler func(Payload) sdk.Result

func (ch Channel) Send(h SendHandler, ctx sdk.Context, msg MsgSend) (result sdk.Result) {
	payload := msg.Payload
	r := ch.runtime(ctx, payload.DatagramType(), msg.DestChain)

	// TODO: check validity of the payload; the module have to be permitted to send the payload
	result = h(msg.Payload)
	if !result.IsOK() {
		return
	}

	data := Datagram{
		Header: Header{
			SrcChain:  ctx.ChainID(),
			DestChain: msg.DestChain,
		},
		Payload: payload,
	}
	r.pushOutgoingQueue(data)

	return
}

type ReceiveHandler func(sdk.Context, Payload) (Payload, sdk.Result)

func (ch Channel) Receive(h ReceiveHandler, ctx sdk.Context, msg MsgReceive) (res sdk.Result) {
	data := msg.Datagram
	payload := data.Payload
	ty := payload.DatagramType()
	chr := ch.runtime(ctx, ty, msg.SrcChain)
	connr := ch.k.runtime(ctx, msg.SrcChain)

	prf := msg.Proof
	destChain := msg.Datagram.Header.DestChain

	if !connr.connEstablished() {
		return ErrConnectionNotEstablished(ch.k.codespace).Result()
	}

	if ctx.ChainID() != destChain {
		return ErrChainMismatch(ch.k.codespace).Result()
	}

	// TODO: verify merkle proof

	seq := chr.getIncomingSequence()
	if seq != prf.Sequence {
		return ErrInvalidSequence(ch.k.codespace).Result()
	}
	chr.setIncomingSequence(seq + 1)

	switch ty {
	case PacketType:
		return receivePacket(h, ctx, chr, data)
	case ReceiptType:
		return receiveReceipt(h, ctx, chr, data)
	default:
		// Source zone sent invalid datagram, reorg needed
		return ErrUnknownDatagramType(ch.k.codespace).Result()
	}
}

func receivePacket(h ReceiveHandler, ctx sdk.Context, r channelRuntime, data Datagram) (res sdk.Result) {
	// Packet handling can fail
	// If fails, reverts all execution done by DatagramHandler

	cctx, write := ctx.CacheContext()
	receipt, res := h(cctx, data.Payload)
	if receipt != nil {
		newdata := Datagram{
			Header:  data.Header.InverseDirection(),
			Payload: receipt,
		}

		r.pushOutgoingQueue(newdata)
	}
	if !res.IsOK() {
		return WrapResult(res)
	}
	write()

	return
}

func receiveReceipt(h ReceiveHandler, ctx sdk.Context, r channelRuntime, data Datagram) (res sdk.Result) {
	// Receipt handling should not fail

	receipt, res := h(ctx, data.Payload)
	if !res.IsOK() {
		panic("IBC Receipt handler should not fail")
	}
	if receipt != nil {
		panic("IBC Receipt handler cannot return new receipt")
	}

	return
}

/*
func cleanup(store sdk.KVStore, cdc *wire.Codec, ty DatagramType, srcChain string) sdk.Result {
	queue := outgoingQueue(store, cdc, ty, srcChain)
}
*/
