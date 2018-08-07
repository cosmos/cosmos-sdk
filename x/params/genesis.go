package params

import (
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

const IrisPrecision = 18
const IrisDenom = "iris"

type CoinConfig struct {
	Denom    string `json:"denom"`
	Precison int64  `json:"precison"`
}

type ParamGenesisState struct {
	Coins []CoinConfig `json:"coins"`
}


func DefaultPrecison(amount int64) sdk.Int {
	return Pow10(IrisPrecision).Mul(sdk.NewInt(amount))
}

func ToBigCoin(ctx sdk.Context, k Getter, coin types.Coin) (types.Coin, error) {
	precison, err := k.GetString(ctx, coin.Denom)
	if err != nil {
		ctx.Logger().Error("module params:%s coin is invalid", coin.Denom)
		return coin, errors.New(fmt.Sprintf("%s coin is invalid", coin.Denom))
	}
	prec,_:= strconv.ParseInt(precison,10,0)
	amount := Pow10(int(prec)).Mul(coin.Amount)
	return types.Coin{Denom: coin.Denom, Amount: amount}, nil
}

func Get(ctx sdk.Context, k Getter, key string, ptr interface{}) error {
	return k.Get(ctx, key, ptr)
}

func DefaultGenesisState() (state ParamGenesisState) {
	state.Coins = append(state.Coins, CoinConfig{
		Denom:    IrisDenom,
		Precison: IrisPrecision,
	})
	return state
}

func InitParamGenesis(ctx sdk.Context, k Keeper, state ParamGenesisState) {
	for _, coin := range state.Coins {
		k.Setter().SetString(ctx, coin.Denom, strconv.FormatInt(coin.Precison, 10))
	}
}

func Pow10(y int) sdk.Int {
	result := sdk.NewInt(1)
	x := sdk.NewInt(10)
	for i := y; i > 0; i >>= 1 {
		if i&1 != 0 {
			result = result.Mul(x)
		}
		x = x.Mul(x)
	}
	return result
}
