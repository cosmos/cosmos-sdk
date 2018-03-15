package ibc

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	//wire "github.com/cosmos/cosmos-sdk/wire"
)

type IBCMapper struct {
	ibcKey sdk.StoreKey

	//	cdc *wire.Codec
}

func NewIBCMapper(ibcKey sdk.StoreKey) IBCMapper {
	//	cdc := wire.NewCodec()

	return IBCMapper{
		ibcKey: ibcKey,
		//		cdc:    cdc,
	}
}

func IngressKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("ingress/%s", srcChain))
}

func EgressKey(destChain string, index int64) []byte {
	return []byte(fmt.Sprintf("egress/%s/%d", destChain, index))
}

func EgressLengthKey(destChain string) []byte {
	return []byte(fmt.Sprintf("egress/%s", destChain))
}

func (ibcm IBCMapper) getEgressLength(store sdk.KVStore, destChain string) int64 {
	bz := store.Get(EgressLengthKey(destChain))
	if bz == nil {
		zero, err := json.Marshal(int64(0)) //ibcm.cdc.MarshalBinary(int64(0))
		if err != nil {
			panic(err)
		}
		store.Set(EgressLengthKey(destChain), zero)
		return 0
	}
	var res int64
	if err := json.Unmarshal(bz, &res); /*ibcm.cdc.UnmarshalBinary(bz, &res)*/ err != nil {
		panic(err)
	}
	return res
}

func (ibcm IBCMapper) GetIngressSequence(ctx sdk.Context, srcChain string) int64 {
	store := ctx.KVStore(ibcm.ibcKey)
	bz := store.Get(IngressKey(srcChain))
	if bz == nil {
		zero, err := json.Marshal(int64(0)) //ibcm.cdc.MarshalBinary(int64(0))
		if err != nil {
			panic(err)
		}
		store.Set(IngressKey(srcChain), zero)
		return 0
	}
	var res int64
	if err := json.Unmarshal(bz, &res); /*ibcm.cdc.UnmarshalBinary(bz, &res)*/ err != nil {
		panic(err)
	}
	return res
}

func (ibcm IBCMapper) SetIngressSequence(ctx sdk.Context, srcChain string, sequence int64) {
	store := ctx.KVStore(ibcm.ibcKey)
	bz, err := json.Marshal(sequence) // ibcm.cdc.MarshalBinary(sequence)
	if err != nil {
		panic(err)
	}
	store.Set(IngressKey(srcChain), bz)
}

func (ibcm IBCMapper) PushPacket(ctx sdk.Context, packet IBCPacket) {
	store := ctx.KVStore(ibcm.ibcKey)
	len := ibcm.getEgressLength(store, packet.DestChain)
	packetbz, err := json.Marshal(packet) // ibcm.cdc.MarshalBinary(packet)
	if err != nil {
		panic(err)
	}
	store.Set(EgressKey(packet.DestChain, len), packetbz)
	lenbz, err := json.Marshal(int64(len + 1)) // ibcm.cdc.MarshalBinary(int64(len + 1))
	if err != nil {
		panic(err)
	}
	store.Set(EgressLengthKey(packet.DestChain), lenbz)
}
