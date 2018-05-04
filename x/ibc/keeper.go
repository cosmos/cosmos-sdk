package ibc

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/tendermint/tendermint/lite"

	"github.com/cosmos/cosmos-sdk/types/lib"
	"github.com/cosmos/cosmos-sdk/wire"
)

type keeper struct {
	key       sdk.StoreKey
	cdc       *wire.Codec
	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) keeper {
	return keeper{
		key:       key,
		cdc:       cdc,
		codespace: codespace,
	}
}

func (keeper keeper) Channel(name string) Channel {
	return Channel{
		keeper: keeper,
		name:   name,
	}
}

type Channel struct {
	keeper keeper
	name   string
}

func (channel Channel) receiveQueue(ctx sdk.Context, chainID string) lib.ListMapper {
	// TODO: check whether the channel is opened or not to the chain
	return lib.NewListMapper(channel.keeper.cdc, channel.keeper.key, "receive/"+chainID)
}

func (channel Channel) receiptQueue(ctx sdk.Context, chainID string) lib.ListMapper {
	// TODO: check whether the channel is opened or not to the chain
	return lib.NewListMapper(channel.keeper.cdc, channel.keeper.key, "receipt/"+chainID)
}

// TODO: Handle invalid IBC packets and return errors.
func (channel Channel) Send(ctx sdk.Context, payload Payload, dest string) sdk.Error {
	if payload.Type() != channel.name {
		return ErrUnauthorizedSend(channel.keeper.codespace)
	}

	// write everything into the state
	packet := Packet{
		Payload:   payload,
		SrcChain:  ctx.ChainID(),
		DestChain: dest,
	}

	queue := channel.receiveQueue(ctx, dest)
	if queue == nil {
		return ErrNoChannelOpened(channel.keeper.codespace, dest)
	}
	queue.Push(ctx, packet)

	return nil
}

func (channel Channel) getSequence(ctx sdk.Context, srcChain string, key []byte) (res int64) {
	keeper := channel.keeper
	store := ctx.KVStore(keeper.key)

	bz := store.Get(key)
	if bz == nil {
		return 0
	}
	unmarshalBinaryPanic(channel.keeper.cdc, bz, &res)
	return
}

func (channel Channel) setSequence(ctx sdk.Context, srcChain string, key []byte, seq int64) {
	keeper := channel.keeper
	store := ctx.KVStore(keeper.key)

	bz := marshalBinaryPanic(keeper.cdc, seq)
	store.Set(key, bz)
}

func (channel Channel) getReceiveSequence(ctx sdk.Context, srcChain string) int64 {
	key := ReceiveSequenceKey(srcChain)
	return channel.getSequence(ctx, srcChain, key)
}

func (channel Channel) setReceiveSequence(ctx sdk.Context, srcChain string, seq int64) {
	key := ReceiveSequenceKey(srcChain)
	channel.setSequence(ctx, srcChain, key, seq)
}

func (channel Channel) getReceiptSequence(ctx sdk.Context, srcChain string) int64 {
	key := ReceiptSequenceKey(srcChain)
	return channel.getSequence(ctx, srcChain, key)
}

func (channel Channel) setReceiptSequence(ctx sdk.Context, srcChain string, seq int64) {
	key := ReceiptSequenceKey(srcChain)
	channel.setSequence(ctx, srcChain, key, seq)
}

/*
// Retrieves the index of the currently stored outgoing IBC packets.
func (keeper keeper) getEgressLength(ctx sdk.Context, destChain string) int64 {
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
*/
/*
func (keeper keeper) getChannelCommit(ctx sdk.Context, srcChain string) (*ValidatorSet, bool) {
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
func (keeper keeper) setChannelCommit(ctx sdk.Context, srcChain string, commit *ValidatorSet) {
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

func (keeper keeper) hasCommit(ctx sdk.Context, srcChain string, height int64) bool {
	store := ctx.KVStore(keeper.key)
	return store.Has(CommitByHeightKey(srcChain, height))
}

func (keeper keeper) setCommit(ctx sdk.Context, srcChain string, height int64, commit lite.FullCommit) {
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

func (keeper keeper) getCommit(ctx sdk.Context, srcChain string, height int64) (commit lite.FullCommit, ok bool) {
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

func (keeper keeper) getCommitHeight(ctx sdk.Context, srcChain string) (res int64, err sdk.Error) {
	store := ctx.KVStore(keeper.key)
	bz := store.Get(CommitHeightKey(srcChain))

	if bz == nil {
		return -1, ErrNoChannelOpened(keeper.codespace, srcChain)
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

func ReceiveSequenceKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("sequence/receive/%s", srcChain))
}

func ReceiptSequenceKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("sequence/receipt/%s", srcChain))
}

func CommitByHeightKey(srcChain string, height int64) []byte {
	return []byte(fmt.Sprintf("commit/height/%s/%d", srcChain, height))
}

func CommitHeightKey(srcChain string) []byte {
	return []byte(fmt.Sprintf("commit/height/%s", srcChain))
}
