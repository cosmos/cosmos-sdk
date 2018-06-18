package ibc

import (
	"github.com/tendermint/tendermint/lite"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/lib"
	"github.com/cosmos/cosmos-sdk/wire"
)

// ------------------------------------------
// Type Definitions

// Keeper manages conn between chains
type Keeper struct {
	key sdk.StoreKey
	cdc *wire.Codec

	codespace sdk.CodespaceType
}

func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		key: key,
		cdc: cdc,

		codespace: codespace,
	}
}

// -----------------------------------------
// Store Accessors

func CommitHeightKey(srcChain string) []byte {
	return append([]byte{0x00}, []byte(srcChain)...)
}

func commitHeight(store sdk.KVStore, cdc *wire.Codec, srcChain string) lib.Value {
	return lib.NewValue(store, cdc, CommitHeightKey(srcChain))
}

func CommitListPrefix(srcChain string) []byte {
	return append([]byte{0x01}, []byte(srcChain)...)
}

func commitList(store sdk.KVStore, cdc *wire.Codec, srcChain string) lib.List {
	return lib.NewList(cdc, store.Prefix(CommitListPrefix(srcChain)), nil)
}

// --------------------------------------
// Keeper Runtime

type connRuntime struct {
	k       Keeper
	height  lib.Value
	commits lib.List
}

func (k Keeper) runtime(ctx sdk.Context, srcChain string) connRuntime {
	store := ctx.KVStore(k.key)
	return connRuntime{
		k:       k,
		height:  commitHeight(store, k.cdc, srcChain),
		commits: commitList(store, k.cdc, srcChain),
	}
}

func (r connRuntime) connEstablished() (established bool) {
	return r.height.Has()
}

func (r connRuntime) getCommitHeight() (height uint64) {
	r.height.MustGet(&height)
	return
}

func (r connRuntime) setCommitHeight(height uint64) {
	r.height.Set(height)
}

func (r connRuntime) getCommit(height uint64) (commit lite.FullCommit) {
	err := r.commits.Get(height, &commit)
	if err != nil {
		panic(err)
	}
	return
}

func (r connRuntime) setCommit(height uint64, commit lite.FullCommit) {
	r.commits.Set(height, commit)
}
