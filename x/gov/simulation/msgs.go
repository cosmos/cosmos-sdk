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
		// TODO random deposit
		deposit := sdk.Coins{sdk.NewCoin("steak", 5)}
		msg := gov.NewMsgSubmitProposal(
			simulation.RandStringOfLength(r, 5),
			simulation.RandStringOfLength(r, 5),
			gov.ProposalTypeText,
			addr,
			deposit,
		)
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		pool := sk.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Sub(sdk.NewRat(5))
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
		lastProposalID := k.GetLastProposalID(ctx)
		if lastProposalID < 1 {
			return "no-operation", nil
		}
		proposalID := int64(r.Intn(int(lastProposalID)))
		msg := gov.NewMsgDeposit(addr, proposalID, deposit)
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
		lastProposalID := k.GetLastProposalID(ctx)
		if lastProposalID < 1 {
			return "no-operation", nil
		}
		proposalID := int64(r.Intn(int(lastProposalID)))
		option := randomVotingOption(r)
		msg := gov.NewMsgVote(addr, proposalID, option)
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

// Pick a random voting option
func randomVotingOption(r *rand.Rand) gov.VoteOption {
	switch r.Intn(4) {
	case 0:
		return gov.OptionYes
	case 1:
		return gov.OptionAbstain
	case 2:
		return gov.OptionNo
	case 3:
		return gov.OptionNoWithVeto
	}
	panic("should not happen")
}
