package pow

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// module users must specify coin denomination and reward (constant) per PoW solution
type PowConfig struct {
	Denomination string
	Reward       int64
}

func NewPowConfig(denomination string, reward int64) PowConfig {
	return PowConfig{denomination, reward}
}

func NewHandler(ck bank.CoinKeeper, pm Mapper, config PowConfig) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MineMsg:
			return handleMineMsg(ctx, ck, pm, config, msg)
		default:
			errMsg := "Unrecognized pow Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMineMsg(ctx sdk.Context, ck bank.CoinKeeper, pm Mapper, config PowConfig, msg MineMsg) sdk.Result {

	// precondition: msg has passed ValidateBasic

	// will this function always be applied atomically?

	lastDifficulty, err := pm.GetLastDifficulty(ctx)
	if err != nil {
		return ErrNonexistentDifficulty().Result()
	}

	newDifficulty := lastDifficulty + 1

	lastCount, err := pm.GetLastCount(ctx)
	if err != nil {
		return ErrNonexistentCount().Result()
	}

	newCount := lastCount + 1

	if msg.Count != newCount {
		return ErrInvalidCount(fmt.Sprintf("invalid count: was %d, should have been %d", msg.Count, newCount)).Result()
	}

	if msg.Difficulty != newDifficulty {
		return ErrInvalidDifficulty(fmt.Sprintf("invalid difficulty: was %d, should have been %d", msg.Difficulty, newDifficulty)).Result()
	}

	_, ckErr := ck.AddCoins(ctx, msg.Sender, []sdk.Coin{sdk.Coin{config.Denomination, config.Reward}})
	if ckErr != nil {
		return ckErr.Result()
	}

	pm.SetLastDifficulty(ctx, newDifficulty)
	pm.SetLastCount(ctx, newCount)

	return sdk.Result{}
}
