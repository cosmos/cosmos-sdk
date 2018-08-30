package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// SimulateMsgUnjail
func SimulateMsgUnjail(k slashing.Keeper) simulation.Operation {
	return func(tb testing.TB, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, fOp []simulation.FutureOperation, err sdk.Error) {
		key := simulation.RandomKey(r, keys)
		address := sdk.AccAddress(key.PubKey().Address())
		msg := slashing.NewMsgUnjail(address)
		if msg.ValidateBasic() != nil {
			tb.Fatalf("expected msg to pass ValidateBasic: %s, log %s", msg.GetSignBytes(), log)
		}
		ctx, write := ctx.CacheContext()
		result := slashing.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("slashing/MsgUnjail/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgUnjail: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil, nil
	}
}
