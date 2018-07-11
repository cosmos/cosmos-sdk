package pow

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

// module users must specify coin denomination and reward (constant) per PoW solution
type Config struct {
	Denomination string
	Reward       int64
}

// genesis info must specify starting difficulty and starting count
type Genesis struct {
	Difficulty uint64 `json:"difficulty"`
	Count      uint64 `json:"count"`
}

// POW Keeper
type Keeper struct {
	key       sdk.StoreKey
	config    Config
	ck        bank.Keeper
	codespace sdk.CodespaceType
}

func NewConfig(denomination string, reward int64) Config {
	return Config{denomination, reward}
}

func NewKeeper(key sdk.StoreKey, config Config, ck bank.Keeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{key, config, ck, codespace}
}

// InitGenesis for the POW module
func InitGenesis(ctx sdk.Context, k Keeper, genesis Genesis) error {
	k.SetLastDifficulty(ctx, genesis.Difficulty)
	k.SetLastCount(ctx, genesis.Count)
	return nil
}

// WriteGenesis for the PoW module
func WriteGenesis(ctx sdk.Context, k Keeper) Genesis {
	difficulty, err := k.GetLastDifficulty(ctx)
	if err != nil {
		panic(err)
	}
	count, err := k.GetLastCount(ctx)
	if err != nil {
		panic(err)
	}
	return Genesis{
		difficulty,
		count,
	}
}

var lastDifficultyKey = []byte("lastDifficultyKey")

// get the last mining difficulty
func (k Keeper) GetLastDifficulty(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.key)
	stored := store.Get(lastDifficultyKey)
	if stored == nil {
		panic("no stored difficulty")
	} else {
		return strconv.ParseUint(string(stored), 0, 64)
	}
}

// set the last mining difficulty
func (k Keeper) SetLastDifficulty(ctx sdk.Context, diff uint64) {
	store := ctx.KVStore(k.key)
	store.Set(lastDifficultyKey, []byte(strconv.FormatUint(diff, 16)))
}

var countKey = []byte("count")

// get the last count
func (k Keeper) GetLastCount(ctx sdk.Context) (uint64, error) {
	store := ctx.KVStore(k.key)
	stored := store.Get(countKey)
	if stored == nil {
		panic("no stored count")
	} else {
		return strconv.ParseUint(string(stored), 0, 64)
	}
}

// set the last count
func (k Keeper) SetLastCount(ctx sdk.Context, count uint64) {
	store := ctx.KVStore(k.key)
	store.Set(countKey, []byte(strconv.FormatUint(count, 16)))
}

// Is the keeper state valid?
func (k Keeper) CheckValid(ctx sdk.Context, difficulty uint64, count uint64) (uint64, uint64, sdk.Error) {

	lastDifficulty, err := k.GetLastDifficulty(ctx)
	if err != nil {
		return 0, 0, ErrNonexistentDifficulty(k.codespace)
	}

	newDifficulty := lastDifficulty + 1

	lastCount, err := k.GetLastCount(ctx)
	if err != nil {
		return 0, 0, ErrNonexistentCount(k.codespace)
	}

	newCount := lastCount + 1

	if count != newCount {
		return 0, 0, ErrInvalidCount(k.codespace, fmt.Sprintf("invalid count: was %d, should have been %d", count, newCount))
	}
	if difficulty != newDifficulty {
		return 0, 0, ErrInvalidDifficulty(k.codespace, fmt.Sprintf("invalid difficulty: was %d, should have been %d", difficulty, newDifficulty))
	}
	return newDifficulty, newCount, nil
}

// Add some coins for a POW well done
func (k Keeper) ApplyValid(ctx sdk.Context, sender sdk.AccAddress, newDifficulty uint64, newCount uint64) sdk.Error {
	_, _, ckErr := k.ck.AddCoins(ctx, sender, []sdk.Coin{sdk.NewCoin(k.config.Denomination, k.config.Reward)})
	if ckErr != nil {
		return ckErr
	}
	k.SetLastDifficulty(ctx, newDifficulty)
	k.SetLastCount(ctx, newCount)
	return nil
}
