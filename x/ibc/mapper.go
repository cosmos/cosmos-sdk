package ibc

import (
	"fmt"

	codec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IBC Mapper
type Mapper struct {
	key       sdk.StoreKey
	cdc       *codec.Codec
	codespace sdk.CodespaceType
}

// XXX: The Mapper should not take a CoinKeeper. Rather have the CoinKeeper
// take an Mapper.
func NewMapper(cdc *codec.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Mapper {
	// XXX: How are these codecs supposed to work?
	return Mapper{
		key:       key,
		cdc:       cdc,
		codespace: codespace,
	}
}

// XXX: This is not the public API. This will change in MVP2 and will henceforth
// only be invoked from another module directly and not through a user
// transaction.
// TODO: Handle invalid IBC packets and return errors.
func (ibcm Mapper) PostIBCPacket(ctx sdk.Context, packet IBCPacket) sdk.Error {
	// write everything into the state
	store := ctx.KVStore(ibcm.key)
	index := ibcm.getEgressLength(store, packet.DestChain)
	bz, err := ibcm.cdc.MarshalBinaryLengthPrefixed(packet)
	if err != nil {
		panic(err)
	}

	store.Set(EgressKey(packet.DestChain, index), bz)
	bz, err = ibcm.cdc.MarshalBinaryLengthPrefixed(index + 1)
	if err != nil {
		panic(err)
	}
	store.Set(EgressLengthKey(packet.DestChain), bz)

	return nil
}

// XXX: In the future every module is able to register it's own handler for
// handling it's own IBC packets. The "ibc" handler will only route the packets
// to the appropriate callbacks.
// XXX: For now this handles all interactions with the CoinKeeper.
// XXX: This needs to do some authentication checking.
func (ibcm Mapper) ReceiveIBCPacket(ctx sdk.Context, packet IBCPacket) sdk.Error {
	return nil
}

// --------------------------
// Functions for accessing the underlying KVStore.

func marshalBinaryPanic(cdc *codec.Codec, value interface{}) []byte {
	res, err := cdc.MarshalBinaryLengthPrefixed(value)
	if err != nil {
		panic(err)
	}
	return res
}

func unmarshalBinaryPanic(cdc *codec.Codec, bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// TODO add description
func (ibcm Mapper) GetIngressSequence(ctx sdk.Context, srcChain string) uint64 {
	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey(srcChain)

	bz := store.Get(key)
	if bz == nil {
		zero := marshalBinaryPanic(ibcm.cdc, int64(0))
		store.Set(key, zero)
		return 0
	}

	var res uint64
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res
}

// TODO add description
func (ibcm Mapper) SetIngressSequence(ctx sdk.Context, srcChain string, sequence uint64) {
	store := ctx.KVStore(ibcm.key)
	key := IngressSequenceKey(srcChain)

	bz := marshalBinaryPanic(ibcm.cdc, sequence)
	store.Set(key, bz)
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (ibcm Mapper) getEgressLength(store sdk.KVStore, destChain string) uint64 {
	bz := store.Get(EgressLengthKey(destChain))
	if bz == nil {
		zero := marshalBinaryPanic(ibcm.cdc, int64(0))
		store.Set(EgressLengthKey(destChain), zero)
		return 0
	}
	var res uint64
	unmarshalBinaryPanic(ibcm.cdc, bz, &res)
	return res
}

// Stores an outgoing IBC packet under "egress/chain_id/index".
func EgressKey(destChain string, index uint64) []byte {
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
