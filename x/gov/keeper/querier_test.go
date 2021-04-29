package keeper_test

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier) (types.DepositParams, types.VotingParams, types.TallyParams) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryParams, types.ParamDeposit}, "/"),
		Data: []byte{},
	}

	bz, err := querier(ctx, []string{types.QueryParams, types.ParamDeposit}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var depositParams types.DepositParams
	require.NoError(t, cdc.UnmarshalJSON(bz, &depositParams))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryParams, types.ParamVoting}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{types.QueryParams, types.ParamVoting}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var votingParams types.VotingParams
	require.NoError(t, cdc.UnmarshalJSON(bz, &votingParams))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryParams, types.ParamTallying}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{types.QueryParams, types.ParamTallying}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var tallyParams types.TallyParams
	require.NoError(t, cdc.UnmarshalJSON(bz, &tallyParams))

	return depositParams, votingParams, tallyParams
}

func getQueriedProposals(
	t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier,
	depositor, voter sdk.AccAddress, status types.ProposalStatus, page, limit int,
) []types.Proposal {

	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposals}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryProposalsParams(page, limit, status, voter, depositor)),
	}

	bz, err := querier(ctx, []string{types.QueryProposals}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var proposals types.Proposals
	require.NoError(t, cdc.UnmarshalJSON(bz, &proposals))

	return proposals
}

func getQueriedDeposit(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, proposalID uint64, depositor sdk.AccAddress) types.Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDeposit}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryDepositParams(proposalID, depositor)),
	}

	bz, err := querier(ctx, []string{types.QueryDeposit}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var deposit types.Deposit
	require.NoError(t, cdc.UnmarshalJSON(bz, &deposit))

	return deposit
}

func getQueriedDeposits(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, proposalID uint64) []types.Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDeposits}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{types.QueryDeposits}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var deposits []types.Deposit
	require.NoError(t, cdc.UnmarshalJSON(bz, &deposits))

	return deposits
}

func getQueriedVote(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, proposalID uint64, voter sdk.AccAddress) types.Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryVoteParams(proposalID, voter)),
	}

	bz, err := querier(ctx, []string{types.QueryVote}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var vote types.Vote
	require.NoError(t, cdc.UnmarshalJSON(bz, &vote))

	return vote
}

func getQueriedVotes(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier,
	proposalID uint64, page, limit int) []types.Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(types.NewQueryProposalVotesParams(proposalID, page, limit)),
	}

	bz, err := querier(ctx, []string{types.QueryVotes}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var votes []types.Vote
	require.NoError(t, cdc.UnmarshalJSON(bz, &votes))

	return votes
}

func TestQueries(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	legacyQuerierCdc := app.LegacyAmino()
	querier := keeper.NewQuerier(app.GovKeeper, legacyQuerierCdc)

	TestAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(20000001))

	oneCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
	consCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10)))

	tp := TestProposal

	depositParams, _, _ := getQueriedParams(t, ctx, legacyQuerierCdc, querier)

	// TestAddrs[0] proposes (and deposits) proposals #1 and #2
	proposal1, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	deposit1 := types.NewDeposit(proposal1.ProposalId, TestAddrs[0], oneCoins)
	depositer1, err := sdk.AccAddressFromBech32(deposit1.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit1.ProposalId, depositer1, deposit1.Amount)
	require.NoError(t, err)

	proposal1.TotalDeposit = proposal1.TotalDeposit.Add(deposit1.Amount...)

	proposal2, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	deposit2 := types.NewDeposit(proposal2.ProposalId, TestAddrs[0], consCoins)
	depositer2, err := sdk.AccAddressFromBech32(deposit2.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit2.ProposalId, depositer2, deposit2.Amount)
	require.NoError(t, err)

	proposal2.TotalDeposit = proposal2.TotalDeposit.Add(deposit2.Amount...)

	// TestAddrs[1] proposes (and deposits) on proposal #3
	proposal3, err := app.GovKeeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	deposit3 := types.NewDeposit(proposal3.ProposalId, TestAddrs[1], oneCoins)
	depositer3, err := sdk.AccAddressFromBech32(deposit3.Depositor)
	require.NoError(t, err)

	_, err = app.GovKeeper.AddDeposit(ctx, deposit3.ProposalId, depositer3, deposit3.Amount)
	require.NoError(t, err)

	proposal3.TotalDeposit = proposal3.TotalDeposit.Add(deposit3.Amount...)

	// TestAddrs[1] deposits on proposals #2 & #3
	deposit4 := types.NewDeposit(proposal2.ProposalId, TestAddrs[1], depositParams.MinDeposit)
	depositer4, err := sdk.AccAddressFromBech32(deposit4.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit4.ProposalId, depositer4, deposit4.Amount)
	require.NoError(t, err)

	proposal2.TotalDeposit = proposal2.TotalDeposit.Add(deposit4.Amount...)
	proposal2.Status = types.StatusVotingPeriod
	proposal2.VotingEndTime = proposal2.VotingEndTime.Add(types.DefaultPeriod)

	deposit5 := types.NewDeposit(proposal3.ProposalId, TestAddrs[1], depositParams.MinDeposit)
	depositer5, err := sdk.AccAddressFromBech32(deposit5.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit5.ProposalId, depositer5, deposit5.Amount)
	require.NoError(t, err)

	proposal3.TotalDeposit = proposal3.TotalDeposit.Add(deposit5.Amount...)
	proposal3.Status = types.StatusVotingPeriod
	proposal3.VotingEndTime = proposal3.VotingEndTime.Add(types.DefaultPeriod)
	// total deposit of TestAddrs[1] on proposal #3 is worth the proposal deposit + individual deposit
	deposit5.Amount = deposit5.Amount.Add(deposit3.Amount...)

	// check deposits on proposal1 match individual deposits

	deposits := getQueriedDeposits(t, ctx, legacyQuerierCdc, querier, proposal1.ProposalId)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit1, deposits[0])

	deposit := getQueriedDeposit(t, ctx, legacyQuerierCdc, querier, proposal1.ProposalId, TestAddrs[0])
	require.Equal(t, deposit1, deposit)

	// check deposits on proposal2 match individual deposits
	deposits = getQueriedDeposits(t, ctx, legacyQuerierCdc, querier, proposal2.ProposalId)
	require.Len(t, deposits, 2)
	// NOTE order of deposits is determined by the addresses
	require.Equal(t, deposit2, deposits[0])
	require.Equal(t, deposit4, deposits[1])

	// check deposits on proposal3 match individual deposits
	deposits = getQueriedDeposits(t, ctx, legacyQuerierCdc, querier, proposal3.ProposalId)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit5, deposits[0])

	deposit = getQueriedDeposit(t, ctx, legacyQuerierCdc, querier, proposal3.ProposalId, TestAddrs[1])
	require.Equal(t, deposit5, deposit)

	// Only proposal #1 should be in types.Deposit Period
	proposals := getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, nil, types.StatusDepositPeriod, 1, 0)
	require.Len(t, proposals, 1)
	require.Equal(t, proposal1, proposals[0])

	// Only proposals #2 and #3 should be in Voting Period
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, nil, types.StatusVotingPeriod, 1, 0)
	require.Len(t, proposals, 2)
	require.Equal(t, proposal2, proposals[0])
	require.Equal(t, proposal3, proposals[1])

	// Addrs[0] votes on proposals #2 & #3
	vote1 := types.NewVote(proposal2.ProposalId, TestAddrs[0], types.OptionYes)
	vote2 := types.NewVote(proposal3.ProposalId, TestAddrs[0], types.OptionYes)
	app.GovKeeper.SetVote(ctx, vote1)
	app.GovKeeper.SetVote(ctx, vote2)

	// Addrs[1] votes on proposal #3
	vote3 := types.NewVote(proposal3.ProposalId, TestAddrs[1], types.OptionYes)
	app.GovKeeper.SetVote(ctx, vote3)

	// Test query voted by TestAddrs[0]
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, TestAddrs[0], types.StatusNil, 1, 0)
	require.Equal(t, proposal2, proposals[0])
	require.Equal(t, proposal3, proposals[1])

	// Test query votes on types.Proposal 2
	votes := getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal2.ProposalId, 1, 0)
	require.Len(t, votes, 1)
	require.Equal(t, vote1, votes[0])

	vote := getQueriedVote(t, ctx, legacyQuerierCdc, querier, proposal2.ProposalId, TestAddrs[0])
	require.Equal(t, vote1, vote)

	// Test query votes on types.Proposal 3
	votes = getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal3.ProposalId, 1, 0)
	require.Len(t, votes, 2)
	require.Equal(t, vote2, votes[0])
	require.Equal(t, vote3, votes[1])

	// Test query all proposals
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, nil, types.StatusNil, 1, 0)
	require.Equal(t, proposal1, proposals[0])
	require.Equal(t, proposal2, proposals[1])
	require.Equal(t, proposal3, proposals[2])

	// Test query voted by TestAddrs[1]
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, TestAddrs[1], types.StatusNil, 1, 0)
	require.Equal(t, proposal3.ProposalId, proposals[0].ProposalId)

	// Test query deposited by TestAddrs[0]
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, TestAddrs[0], nil, types.StatusNil, 1, 0)
	require.Equal(t, proposal1.ProposalId, proposals[0].ProposalId)

	// Test query deposited by addr2
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, TestAddrs[1], nil, types.StatusNil, 1, 0)
	require.Equal(t, proposal2.ProposalId, proposals[0].ProposalId)
	require.Equal(t, proposal3.ProposalId, proposals[1].ProposalId)

	// Test query voted AND deposited by addr1
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, TestAddrs[0], TestAddrs[0], types.StatusNil, 1, 0)
	require.Equal(t, proposal2.ProposalId, proposals[0].ProposalId)
}

func TestPaginatedVotesQuery(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	legacyQuerierCdc := app.LegacyAmino()

	proposal := types.Proposal{
		ProposalId: 100,
		Status:     types.StatusVotingPeriod,
	}

	app.GovKeeper.SetProposal(ctx, proposal)

	votes := make([]types.Vote, 20)
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	addrMap := make(map[string]struct{})
	genAddr := func() string {
		addr := make(sdk.AccAddress, 20)
		for {
			random.Read(addr)
			addrStr := addr.String()
			if _, ok := addrMap[addrStr]; !ok {
				addrMap[addrStr] = struct{}{}
				return addrStr
			}
		}
	}
	for i := range votes {
		vote := types.Vote{
			ProposalId: proposal.ProposalId,
			Voter:      genAddr(),
			Option:     types.OptionYes,
		}
		votes[i] = vote
		app.GovKeeper.SetVote(ctx, vote)
	}

	querier := keeper.NewQuerier(app.GovKeeper, legacyQuerierCdc)

	// keeper preserves consistent order for each query, but this is not the insertion order
	all := getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal.ProposalId, 1, 0)
	require.Equal(t, len(all), len(votes))

	type testCase struct {
		description string
		page        int
		limit       int
		votes       []types.Vote
	}
	for _, tc := range []testCase{
		{
			description: "SkipAll",
			page:        2,
			limit:       len(all),
		},
		{
			description: "GetFirstChunk",
			page:        1,
			limit:       10,
			votes:       all[:10],
		},
		{
			description: "GetSecondsChunk",
			page:        2,
			limit:       10,
			votes:       all[10:],
		},
		{
			description: "InvalidPage",
			page:        -1,
		},
	} {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			votes := getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal.ProposalId, tc.page, tc.limit)
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}
