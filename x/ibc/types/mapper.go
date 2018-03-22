package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
)

type Sender interface {
	Push(sdk.Context, Payload, string)
}

type sender struct {
	keeper  keeper
	allowed map[string]bool
}

// XXX: This is not the public API. This will change in MVP2 and will henceforth
// only be invoked from another module directly and not through a user
// transaction.
// TODO: Handle invalid IBC packets and return errors.
func (sender sender) push(ctx sdk.Context, payload Payload, dest string) sdk.Error {
	// write everything into the state
	store := ctx.KVStore(sender.keeper.key)
	packet := Packet{
		Payload:   payload,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}
	keeper := sender.keeper
	index := keeper.getEgressLength(store, dest)
	bz, err := keeper.cdc.MarshalBinary(packet)
	if err != nil {
		panic(err)
	}

	store.Set(EgressKey(dest, index), bz)
	bz, err = keeper.cdc.MarshalBinary(int64(index + 1))
	if err != nil {
		panic(err)
	}
	store.Set(EgressLengthKey(dest), bz)

	return nil
}

func (sender sender) Push(ctx sdk.Context, payload Payload, dest string) {
	_, ok := sender.allowed[payload.Type()]
	if !ok {
		panic("")
	}
	sender.push(ctx, payload, dest)
}

type keeper struct {
	key      sdk.StoreKey
	cdc      *wire.Codec
	dispatch map[string]Handler
}

func (keeper keeper) Sender(payloads ...Payload) Sender {
	allowed := make(map[string]bool)
	for _, payload := range payloads {
		allowed[payload.Type()] = true
	}
	return sender{
		keeper:  keeper,
		allowed: allowed,
	}
}

func (keeper keeper) RegisterHandler(name string, handler Handler) {
	keeper.dispatch[name] = handler
}

func (keeper keeper) Receive(ctx sdk.Context, packet Packet, seq int64) sdk.Error {
	if packet.DestChain != ctx.ChainID() {
		// TODO: route?
		return ErrChainMismatch()
	}

	expected := keeper.getIngressSequence(ctx, packet.SrcChain)
	if seq != expected {
		return ErrInvalidSequence()
	}

	payload := packet.Payload
	res := keeper.dispatch[payload.Type()](ctx, payload)

	keeper.setIngressSequence(ctx, packet.SrcChain, seq+1)

	return res
}

type Keeper interface {
	Sender(...Payload) Sender
	RegisterHandler(string, Handler)
	Receive(sdk.Context, Packet, int64) sdk.Error
}

// XXX: The Keeper should not take a CoinKeeper. Rather have the CoinKeeper
// take an Keeper.
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey) Keeper {
	// XXX: How are these codecs supposed to work?
	return keeper{
		key:      key,
		cdc:      cdc,
		dispatch: make(map[string]Handler),
	}
}

// XXX: In the future every module is able to register it's own handler for
// handling it's own IBC packets. The "ibc" handler will only route the packets
// to the appropriate callbacks.
// XXX: For now this handles all interactions with the CoinKeeper.
// XXX: This needs to do some authentication checking.
func (keeper keeper) ReceiveIBCPacket(ctx sdk.Context, packet Packet) sdk.Error {
	return nil
}

// --------------------------
// Functions for accessing the underlying KVStore.

func marshalBinaryPanic(cdc *wire.Codec, value interface{}) []byte {
	res, err := cdc.MarshalBinary(value)
	if err != nil {
		panic(err)
	}
	return res
}

func unmarshalBinaryPanic(cdc *wire.Codec, bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinary(bz, ptr)
	if err != nil {
		panic(err)
	}
}

func (keeper keeper) getIngressSequence(ctx sdk.Context, srcChain string) int64 {
	store := ctx.KVStore(keeper.key)
	key := IngressSequenceKey(srcChain)

	bz := store.Get(key)
	if bz == nil {
		zero := marshalBinaryPanic(keeper.cdc, int64(0))
		store.Set(key, zero)
		return 0
	}

	var res int64
	unmarshalBinaryPanic(keeper.cdc, bz, &res)
	return res
}

func (keeper keeper) setIngressSequence(ctx sdk.Context, srcChain string, sequence int64) {
	store := ctx.KVStore(keeper.key)
	key := IngressSequenceKey(srcChain)

	bz := marshalBinaryPanic(keeper.cdc, sequence)
	store.Set(key, bz)
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (keeper keeper) getEgressLength(store sdk.KVStore, destChain string) int64 {
	bz := store.Get(EgressLengthKey(destChain))
	if bz == nil {
		zero := marshalBinaryPanic(keeper.cdc, int64(0))
		store.Set(EgressLengthKey(destChain), zero)
		return 0
	}
	var res int64
	unmarshalBinaryPanic(keeper.cdc, bz, &res)
	return res
}

// Stores an outgoing IBC packet under "egress/chain_id/index".
func EgressKey(destChain string, index int64) []byte {
	return []byte(fmt.Sprintf("egress/%s/%d", destChain, index))
}

// Stores the number of outgoing IBC packets under "egress/index".
func EgressLengthKey(destChain string) []byte {
	return []byte(fmt.Sprintf("egress/%s", destChain))
}

// Stores the sequence number of incoming IBC packet under "ingress/index".
func IngressSequenceKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("ingress/%s", srcChain))
}
