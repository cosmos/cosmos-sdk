package params_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// Copied from gov and modified

func getMockApp(t *testing.T, numGenAccs int) (*mock.App, params.ProposalKeeper, bank.BaseKeeper, staking.Keeper, gov.Keeper, []sdk.AccAddress) {
	mapp := mock.NewApp()

	staking.RegisterCodec(mapp.Cdc)

	keyStaking := sdk.NewKVStoreKey(staking.StoreKey)
	tkeyStaking := sdk.NewTransientStoreKey(staking.TStoreKey)
	keyGov := sdk.NewKVStoreKey(gov.StoreKey)

	pk := mapp.ParamsKeeper
	ck := bank.NewBaseKeeper(mapp.AccountKeeper, mapp.ParamsKeeper.Subspace(bank.DefaultParamspace), bank.DefaultCodespace)
	sk := staking.NewKeeper(mapp.Cdc, keyStaking, tkeyStaking, ck, pk.Subspace(staking.DefaultParamspace), staking.DefaultCodespace)
	gk := gov.NewKeeper(mapp.Cdc, keyGov, pk, pk.Subspace("testgov"), ck, sk, gov.DefaultCodespace)
	keeper := params.NewProposalKeeperWithExisting(mapp.ParamsKeeper, gk)
	gk.Router().AddRoute(params.RouteKey, params.NewProposalHandler(keeper))

	mapp.Router().AddRoute(params.RouteKey, params.NewHandler(keeper))

	require.NoError(t, mapp.CompleteSetup(keyStaking, tkeyStaking, keyGov))

	valTokens := sdk.TokensFromTendermintPower(42)
	genAccs, addrs, pubKeys, privKeys := mock.CreateGenAccounts(numGenAccs,
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, valTokens)})

	mock.SetGenesis(mapp, genAccs)

	return mapp, keeper, ck, sk, gk, addrs
}

var pubkeys = []crypto.PubKey{ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey(), ed25519.GenPrivKey().PubKey()}

func createValidators(t *testing.T, stakingHandler sdk.Handler, ctx sdk.Context, addrs []sdk.ValAddress, powerAmt []int64) {
	require.True(t, len(addrs) <= len(pubkeys), "Not enough pubkeys specified at top of file.")

	for i := 0; i < len(addrs); i++ {

		valTokens := sdk.TokensFromTendermintPower(powerAmt[i])
		valCreateMsg := staking.NewMsgCreateValidator(
			addrs[i], pubkeys[i], sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
			staking.NewDescription("T", "E", "S", "T"),
			staking.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			sdk.OneInt(),
		)

		res := stakingHandler(ctx, valCreateMsg)
		require.True(t, res.IsOK())
	}
}

func TestProposalPassedEndblocker(t *testing.T) {
	mapp, keeper, ck, sk, gk, addrs := getMockApp(t, 10)
	gov.SortAddresses(addrs)
	header := abci.Header{Height: mapp.LastBlockHeight() + 1}
	mapp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})
	ck.SetSendEnabled(ctx, true)
	stakingHandler := staking.NewHandler(sk)
	govHandler := gov.NewHandler(gk)

	valAddrs := make([]sdk.ValAddress, 2)
	valAddrs[0], valAddrs[1] = sdk.ValAddress(addrs[0]), sdk.ValAddress(addrs[1])

	createValidators(t, stakingHandler, ctx, valAddrs, []int64{5, 5})
	staking.EndBlocker(ctx, sk)

	proposal, err := testSubmitProposal(ctx, keeper, tp)
	require.NoError(t, err)
	proposalID := proposal.ProposalID
	proposal.Status = StatusDepositPeriod
	keeper.SetProposal(ctx, proposal)

	proposalCoins := sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromTendermintPower(10))}
	newDepositMsg := NewMsgDeposit(addrs[0], proposalID, proposalCoins)
	res := govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())
	newDepositMsg = NewMsgDeposit(addrs[1], proposalID, proposalCoins)
	res = govHandler(ctx, newDepositMsg)
	require.True(t, res.IsOK())

	err = keeper.AddVote(ctx, proposalID, addrs[0], OptionYes)
	require.NoError(t, err)
	err = keeper.AddVote(ctx, proposalID, addrs[1], OptionYes)
	require.NoError(t, err)

	newHeader := ctx.BlockHeader()
	newHeader.Time = ctx.BlockHeader().Time.Add(keeper.GetDepositParams(ctx).MaxDepositPeriod).Add(keeper.GetVotingParams(ctx).VotingPeriod)
	ctx = ctx.WithBlockHeader(newHeader)

	// Handler registered, not panics
	var restags sdk.Tags
	require.NotPanics(t, func() { restags = EndBlocker(ctx, keeper) })

	require.Equal(t, sdk.MakeTag(tags.ProposalResult, tags.ActionProposalPassed), restags[1])
}
