package pow

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Mapper struct {
	key sdk.StoreKey
}

func NewMapper(key sdk.StoreKey) Mapper {
	return Mapper{key}
}

var lastDifficultyKey = []byte("lastDifficultyKey")

func (pm Mapper) GetLastDifficulty(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(pm.key)
	stored := store.Get(lastDifficultyKey)
	if stored == nil {
		// return the default difficulty of 1 if not set
		// this works OK for this module, but a way to initalize the store (a "genesis block" for the module) might be better in general
		return uint64(1), nil
	} else {
		return strconv.ParseUint(string(stored), 0, 64)
	}
}

func (pm Mapper) SetLastDifficulty(ctx sdk.Context, diff uint64) {
	store := ctx.KVStore(pm.key)
	store.Set(lastDifficultyKey, []byte(strconv.FormatUint(diff, 16)))
}

var countKey = []byte("count")

func (pm Mapper) GetLastCount(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(pm.key)
	stored := store.Get(countKey)
	if stored == nil {
		return uint64(0), nil
	} else {
		return strconv.ParseUint(string(stored), 0, 64)
	}
}

func (pm Mapper) SetLastCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(pm.key)
	store.Set(countKey, []byte(strconv.FormatUint(count, 16)))
}
