package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/lite"

	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/x/ibc/types"
)

type keeper struct {
	key        sdk.StoreKey
	cdc        *wire.Codec
	dispatcher types.Dispatcher
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey) keeper {
	return keeper{
		key:        key,
		cdc:        cdc,
		dispatcher: types.NewDispatcher(),
	}
}

// XXX: This is not the public API. This will change in MVP2 and will henceforth
// only be invoked from another module directly and not through a user
// transaction.
// TODO: Handle invalid IBC packets and return errors.
func (sender keeper) Push(ctx sdk.Context, payload types.Payload, dest string) {
	// write everything into the state
	store := ctx.KVStore(sender.key)
	packet := types.Packet{
		Payload:   payload,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}
	index := sender.getEgressLength(store, dest)
	bz, err := sender.cdc.MarshalBinary(packet)
	if err != nil {
		panic(err)
	}

	store.Set(EgressKey(dest, index), bz)
	bz, err = sender.cdc.MarshalBinary(int64(index + 1))
	if err != nil {
		panic(err)
	}
	store.Set(EgressLengthKey(dest), bz)
}

func (keeper keeper) Dispatcher() types.Dispatcher {
	return keeper.dispatcher
}

func (keeper keeper) Sender() types.Sender {
	return keeper
}

func (keeper keeper) Receive(ctx sdk.Context, packet types.Packet) sdk.Error {
	if packet.DestChain != ctx.ChainID() {
		// TODO: route?
		return types.ErrChainMismatch()
	}

	payload := packet.Payload
	res := keeper.dispatcher.Dispatch(payload.Type())(ctx, payload)

	return res
}

/*
func (keeper keeper) getChannelCommit(ctx sdk.Context, srcChain string) (*types.ValidatorSet, bool) {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(ChannelCommitKey(srcChain))
	if bz == nil {
		return *types.ValidatorSet{}, false
	}

	var commit *types.ValidatorSet
	err := keeper.cdc.UnmarshalBinary(bz, &commit)
	if err != nil {
		panic(err)
	}

	return commit, true
}
*/
/*
func (keeper keeper) setChannelCommit(ctx sdk.Context, srcChain string, commit *types.ValidatorSet) {
	store := ctx.KVStore(keeper.key)
	bz, err := keeper.cdc.MarshalBinary(commit)
	if err != nil {
		panic(err)
	}
	store.Set(ChannelCommitKey(srcChain), bz)
}
*/
/*
func (keeper keeper) getCertifier(ctx sdk.Context, srcChain string, height int64) (*lite.Inquiring, bool) {
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

func (keeper keeper) hasChannelCommit(ctx sdk.Context, srcChain string, height int64) bool {
	store := ctx.KVStore(keeper.key)
	return store.Has(CommitByHeightKey(srcChain, height))
}

func (keeper keeper) setChannelCommit(ctx sdk.Context, srcChain string, height int64, commit lite.FullCommit) {
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

func (keeper keeper) getChannelCommit(ctx sdk.Context, srcChain string, height int64) (commit lite.FullCommit) {
	store := ctx.KVStore(keeper.key)

	bz := store.Get(CommitByHeightKey(srcChain, height))
	if err := keeper.cdc.UnmarshalBinary(bz, &commit); err != nil {
		panic(err)
	}

	return
}

func (keeper keeper) getChannelCommitHeight(ctx sdk.Context, srcChain string) (res int64, err sdk.Error) {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(CommitHeightKey(srcChain))

	if bz == nil {
		return -1, types.ErrNoChannelOpened(srcChain)
	}

	if err := keeper.cdc.UnmarshalBinary(bz, &res); err != nil {
		panic(err)
	}

	return
}

func (keeper keeper) getChannelRecentCommit(ctx sdk.Context, srcChain string) (height int64, commit lite.FullCommit, err sdk.Error) {
	height, err = keeper.getChannelCommitHeight(ctx, srcChain)
	if err != nil {
		return
	}

	commit = keeper.getChannelCommit(ctx, srcChain, height)
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
