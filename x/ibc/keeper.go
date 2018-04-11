package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/lite"

	"github.com/cosmos/cosmos-sdk/wire"
)

type keeperFactory struct {
	key sdk.StoreKey
	cdc *wire.Codec
}

func NewKeeperFactory(cdc *wire.Codec, key sdk.StoreKey) keeperFactory {
	return keeperFactory{
		key: key,
		cdc: cdc,
	}
}

func (kf keeperFactory) Port(port string) Keeper {
	return Keeper{
		key:  kf.key,
		cdc:  kf.cdc,
		port: port,
	}
}

type Keeper struct {
	key  sdk.StoreKey
	cdc  *wire.Codec
	port string
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		key: key,
		cdc: cdc,
	}
}

// TODO: Handle invalid IBC packets and return errors.
func (keeper Keeper) Send(ctx sdk.Context, payload Payload, dest string) sdk.Error {
	if payload.Type() != keeper.port {
		return ErrUnauthorizedSend()
	}

	// write everything into the state
	store := ctx.KVStore(keeper.key)
	packet := Packet{
		Payload:   payload,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}
	index := keeper.getEgressLength(ctx, dest)
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

func (keeper Keeper) sendReceipt(ctx sdk.Context, payload Payload, src string) sdk.Error {
	if payload.Type() != keeper.port {
		return ErrUnauthorizedSendReceipt()
	}

	store := ctx.KVStore(keeper.key)

}

/*
func (keeper Keeper) getChannelCommit(ctx sdk.Context, srcChain string) (*ValidatorSet, bool) {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(ChannelCommitKey(srcChain))
	if bz == nil {
		return *ValidatorSet{}, false
	}

	var commit *ValidatorSet
	err := keeper.cdc.UnmarshalBinary(bz, &commit)
	if err != nil {
		panic(err)
	}

	return commit, true
}
*/
/*
func (keeper Keeper) setChannelCommit(ctx sdk.Context, srcChain string, commit *ValidatorSet) {
	store := ctx.KVStore(keeper.key)
	bz, err := keeper.cdc.MarshalBinary(commit)
	if err != nil {
		panic(err)
	}
	store.Set(ChannelCommitKey(srcChain), bz)
}
*/
/*
func (keeper Keeper) getCertifier(ctx sdk.Context, srcChain string, height int64) (*lite.Inquiring, bool) {
	if height <= 0 {
		height = keeper.getChannelHeight(ctx, srcChain)
	}

	commit, ok := keeper.getChannelCommit(ctx, srcChain, height)
	if !ok {
		return nil, false
	}

	cert := lite.NewInquiring(srcChain, commit)
}
*/

func (keeper Keeper) hasChannelCommit(ctx sdk.Context, srcChain string, height int64) bool {
	store := ctx.KVStore(keeper.key)
	return store.Has(CommitByHeightKey(srcChain, height))
}

func (keeper Keeper) setChannelCommit(ctx sdk.Context, srcChain string, height int64, commit lite.FullCommit) {
	store := ctx.KVStore(keeper.key)

	bz, err := keeper.cdc.MarshalBinary(commit)
	if err != nil {
		panic(err)
	}

	store.Set(CommitByHeightKey(srcChain, height), bz)

	bz, err = keeper.cdc.MarshalBinary(height)
	if err != nil {
		panic(err)
	}

	store.Set(CommitHeightKey(srcChain), bz)
}

func (keeper Keeper) getChannelCommit(ctx sdk.Context, srcChain string, height int64) (commit lite.FullCommit, ok bool) {
	store := ctx.KVStore(keeper.key)

	bz := store.Get(CommitByHeightKey(srcChain, height))
	if bz == nil {
		return commit, false
	}

	if err := keeper.cdc.UnmarshalBinary(bz, &commit); err != nil {
		panic(err)
	}

	return commit, true
}

func (keeper Keeper) getChannelCommitHeight(ctx sdk.Context, srcChain string) (res int64, err sdk.Error) {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(CommitHeightKey(srcChain))

	if bz == nil {
		return -1, ErrNoChannelOpened(srcChain)
	}

	if err := keeper.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}

	return
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

func (keeper Keeper) getIngressSequence(ctx sdk.Context, srcChain string) int64 {
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

func (keeper Keeper) setIngressSequence(ctx sdk.Context, srcChain string, sequence int64) {
	store := ctx.KVStore(keeper.key)
	key := IngressSequenceKey(srcChain)

	bz := marshalBinaryPanic(keeper.cdc, sequence)
	store.Set(key, bz)
}

// Retrieves the index of the currently stored outgoing IBC packets.
func (keeper Keeper) getEgressLength(ctx sdk.Context, destChain string) int64 {
	store := ctx.KVStore(keeper.key)
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

func EgressKey(destChain string, index int64) []byte {
	return []byte(fmt.Sprintf("egress/%s/%d", destChain, index))
}

func EgressLengthKey(destChain string) []byte {
	return []byte(fmt.Sprintf("egress/%s", destChain))
}

func IngressSequenceKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("ingress/%s", srcChain))
}

func CommitByHeightKey(srcChain string, height int64) []byte {
	return []byte(fmt.Sprintf("commit/height/%s/%d", srcChain, height))
}

func CommitHeightKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("commit/height/%s", srcChain))
}
