package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

// SimulateMsgUnjail
func SimulateMsgUnjail(k slashing.Keeper) simulation.Operation {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, fOp []simulation.FutureOperation, err sdk.Error) {
		key := simulation.RandomKey(r, keys)
		address := sdk.ValAddress(key.PubKey().Address())
		msg := slashing.NewMsgUnjail(address)
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
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
