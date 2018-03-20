package staking

import (
	abci "github.com/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

func NewHandler(sm StakingMapper, ck bank.CoinKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case BondMsg:
			return handleBondMsg(ctx, sm, ck, msg)
		case UnbondMsg:
			return handleUnbondMsg(ctx, sm, ck, msg)
		default:
			return sdk.ErrUnknownRequest("No match for message type.").Result()
		}
	}
}

func handleBondMsg(ctx sdk.Context, sm StakingMapper, ck bank.CoinKeeper, msg BondMsg) sdk.Result {
	_, err := ck.SubtractCoins(ctx, msg.Address, []sdk.Coin{msg.Stake})
	if err != nil {
		return err.Result()
	}

	power, err := sm.Bond(ctx, msg.Address, msg.PubKey, msg.Stake.Amount)
	if err != nil {
		return err.Result()
	}

	valSet := abci.Validator{
		PubKey: msg.PubKey.Bytes(),
		Power:  power,
	}

	return sdk.Result{
		Code:             sdk.CodeOK,
		ValidatorUpdates: abci.Validators{valSet},
	}
}

func handleUnbondMsg(ctx sdk.Context, sm StakingMapper, ck bank.CoinKeeper, msg UnbondMsg) sdk.Result {
	pubKey, power, err := sm.Unbond(ctx, msg.Address)
	if err != nil {
		return err.Result()
	}

	stake := sdk.Coin{
		Denom:  "mycoin",
		Amount: power,
	}
	_, err = ck.AddCoins(ctx, msg.Address, sdk.Coins{stake})
	if err != nil {
		return err.Result()
	}

	valSet := abci.Validator{
		PubKey: pubKey.Bytes(),
		Power:  int64(0),
	}

	return sdk.Result{
		Code:             sdk.CodeOK,
		ValidatorUpdates: abci.Validators{valSet},
	}
}
