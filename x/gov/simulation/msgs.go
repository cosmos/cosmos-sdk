package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// SimulateMsgSubmitProposal
func SimulateMsgSubmitProposal(k gov.Keeper, sk stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		key := simulation.RandomKey(r, keys)
		addr := sdk.AccAddress(key.PubKey().Address())
		deposit := sdk.Coins{sdk.NewCoin("steak", 10)}
		msg := gov.NewMsgSubmitProposal(
			simulation.RandStringOfLength(r, 10),
			simulation.RandStringOfLength(r, 10),
			gov.ProposalTypeText,
			addr,
			deposit,
		)
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		pool := sk.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Sub(sdk.NewRat(10))
		sk.SetPool(ctx, pool)
		result := gov.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("gov/MsgSubmitProposal/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgSubmitProposal: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgDeposit
func SimulateMsgDeposit(k gov.Keeper, sk stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		key := simulation.RandomKey(r, keys)
		addr := sdk.AccAddress(key.PubKey().Address())
		deposit := sdk.Coins{sdk.NewCoin("steak", 10)}
		msg := gov.NewMsgDeposit(addr, int64(1), deposit)
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := gov.NewHandler(k)(ctx, msg)
		pool := sk.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Sub(sdk.NewRat(10))
		sk.SetPool(ctx, pool)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("gov/MsgDeposit/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgDeposit: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgVote
func SimulateMsgVote(k gov.Keeper, sk stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		key := simulation.RandomKey(r, keys)
		addr := sdk.AccAddress(key.PubKey().Address())
		msg := gov.NewMsgVote(addr, int64(1), gov.OptionYes)
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := gov.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("gov/MsgVote/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgVote: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}
