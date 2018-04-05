package pow

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

// module users must specify coin denomination and reward (constant) per PoW solution
type PowConfig struct {
	Denomination string
	Reward       int64
}

// genesis info must specify starting difficulty and starting count
type PowGenesis struct {
	Difficulty uint64 `json:"difficulty"`
	Count      uint64 `json:"count"`
}

type Keeper struct {
	key    sdk.StoreKey
	config PowConfig
	ck     bank.CoinKeeper
}

func NewPowConfig(denomination string, reward int64) PowConfig {
	return PowConfig{denomination, reward}
}

func NewKeeper(key sdk.StoreKey, config PowConfig, ck bank.CoinKeeper) Keeper {
	return Keeper{key, config, ck}
}

func (pk Keeper) InitGenesis(ctx sdk.Context, genesis PowGenesis) error {
	pk.SetLastDifficulty(ctx, genesis.Difficulty)
	pk.SetLastCount(ctx, genesis.Count)
	return nil
}

var lastDifficultyKey = []byte("lastDifficultyKey")

func (pk Keeper) GetLastDifficulty(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(pk.key)
	stored := store.Get(lastDifficultyKey)
	if stored == nil {
		panic("no stored difficulty")
	} else {
		return strconv.ParseUint(string(stored), 0, 64)
	}
}

func (pk Keeper) SetLastDifficulty(ctx sdk.Context, diff uint64) {
	store := ctx.KVStore(pk.key)
	store.Set(lastDifficultyKey, []byte(strconv.FormatUint(diff, 16)))
}

var countKey = []byte("count")

func (pk Keeper) GetLastCount(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(pk.key)
	stored := store.Get(countKey)
	if stored == nil {
		panic("no stored count")
	} else {
		return strconv.ParseUint(string(stored), 0, 64)
	}
}

func (pk Keeper) SetLastCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(pk.key)
	store.Set(countKey, []byte(strconv.FormatUint(count, 16)))
}

func (pk Keeper) CheckValid(ctx sdk.Context, difficulty uint64, count uint64) (uint64, uint64, sdk.Error) {

	lastDifficulty, err := pk.GetLastDifficulty(ctx)
	if err != nil {
		return 0, 0, ErrNonexistentDifficulty()
	}

	newDifficulty := lastDifficulty + 1

	lastCount, err := pk.GetLastCount(ctx)
	if err != nil {
		return 0, 0, ErrNonexistentCount()
	}

	newCount := lastCount + 1

	if count != newCount {
		return 0, 0, ErrInvalidCount(fmt.Sprintf("invalid count: was %d, should have been %d", count, newCount))
	}

	if difficulty != newDifficulty {
		return 0, 0, ErrInvalidDifficulty(fmt.Sprintf("invalid difficulty: was %d, should have been %d", difficulty, newDifficulty))
	}

	return newDifficulty, newCount, nil

}

func (pk Keeper) ApplyValid(ctx sdk.Context, sender sdk.Address, newDifficulty uint64, newCount uint64) sdk.Error {
	_, ckErr := pk.ck.AddCoins(ctx, sender, []sdk.Coin{sdk.Coin{pk.config.Denomination, pk.config.Reward}})
	if ckErr != nil {
		return ckErr
	}
	pk.SetLastDifficulty(ctx, newDifficulty)
	pk.SetLastCount(ctx, newCount)
	return nil
}
