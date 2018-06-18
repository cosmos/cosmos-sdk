package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
	"github.com/cosmos/cosmos-sdk/wire"
)

// ------------------------------------------
// Type Definitions

type Channel struct {
	k   Keeper
	key sdk.KVStoreGetter
}

func (k Keeper) Channel(key sdk.KVStoreGetter) Channel {
	return Channel{
		k:   k,
		key: key,
	}
}

type DatagramType byte

const (
	PacketType = DatagramType(iota)
	ReceiptType
)

type Header struct {
	SrcChain  string
	DestChain string
}

func (h Header) InverseDirection() Header {
	return Header{
		SrcChain:  h.DestChain,
		DestChain: h.SrcChain,
	}
}

type Payload interface {
	Type() string
	ValidateBasic() sdk.Error
	GetSigners() []sdk.AccAddress
	DatagramType() DatagramType
}

type Datagram struct {
	Header
	// Should we unexport Payload to possible modification from the modules?
	Payload
}

type Proof struct {
	Height   uint64
	Sequence uint64
}

// -------------------------------------------
// Store Accessors

func OutgoingQueuePrefix(ty DatagramType, chainid string) []byte {
	return append(append([]byte{0x00}, byte(ty)), []byte(chainid)...)
}

func outgoingQueue(store sdk.KVStore, cdc *wire.Codec, ty DatagramType, chainid string) lib.Linear {
	return lib.NewLinear(cdc, store.Prefix(OutgoingQueuePrefix(ty, chainid)), nil)
}

func IncomingSequenceKey(ty DatagramType, chainid string) []byte {
	return append(append([]byte{0x01}, byte(ty)), []byte(chainid)...)
}

func incomingSequence(store sdk.KVStore, cdc *wire.Codec, ty DatagramType, chainid string) lib.Value {
	return lib.NewValue(store, cdc, IncomingSequenceKey(ty, chainid))
}

// --------------------------------------------
// Channel Runtime

type channelRuntime struct {
	ch               Channel
	outgoingQueue    lib.Queue
	incomingSequence lib.Value
	thisChain        string
	thatChain        string
}

func (ch Channel) runtime(ctx sdk.Context, ty DatagramType, thatChain string) channelRuntime {
	store := ctx.KVStore(ch.k.key)

	return channelRuntime{
		ch:               ch,
		outgoingQueue:    outgoingQueue(store, ch.k.cdc, ty, thatChain),
		incomingSequence: incomingSequence(store, ch.k.cdc, ty, thatChain),
		thisChain:        ctx.ChainID(),
		thatChain:        thatChain,
	}
}

func (r channelRuntime) pushOutgoingQueue(data Datagram) {
	r.outgoingQueue.Push(data)
}

func (r channelRuntime) getIncomingSequence() (res uint64) {
	ok := r.incomingSequence.Get(&res)
	if !ok {
		return 0
	}
	return
}

func (r channelRuntime) setIncomingSequence(seq uint64) {
	r.incomingSequence.Set(seq)
}
