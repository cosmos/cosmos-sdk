package keeper

import (
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier) (types.DepositParams, types.VotingParams, types.TallyParams) {
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
	t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier,
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

func getQueriedDeposit(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64, depositor sdk.AccAddress) types.Deposit {
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

func getQueriedDeposits(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) []types.Deposit {
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

func getQueriedVote(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64, voter sdk.AccAddress) types.Vote {
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

func getQueriedVotes(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier,
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
	ctx, _, keeper, _, _ := createTestInput(t, false, 1000)
	querier := NewQuerier(keeper)

	oneCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
	consCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10)))

	tp := TestProposal

	depositParams, _, _ := getQueriedParams(t, ctx, keeper.cdc, querier)

	// TestAddrs[0] proposes (and deposits) proposals #1 and #2
	proposal1, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	deposit1 := types.NewDeposit(proposal1.ProposalID, TestAddrs[0], oneCoins)
	_, err = keeper.AddDeposit(ctx, deposit1.ProposalID, deposit1.Depositor, deposit1.Amount)
	require.NoError(t, err)

	proposal1.TotalDeposit = proposal1.TotalDeposit.Add(deposit1.Amount...)

	proposal2, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	deposit2 := types.NewDeposit(proposal2.ProposalID, TestAddrs[0], consCoins)
	_, err = keeper.AddDeposit(ctx, deposit2.ProposalID, deposit2.Depositor, deposit2.Amount)
	require.NoError(t, err)

	proposal2.TotalDeposit = proposal2.TotalDeposit.Add(deposit2.Amount...)

	// TestAddrs[1] proposes (and deposits) on proposal #3
	proposal3, err := keeper.SubmitProposal(ctx, tp)
	require.NoError(t, err)
	deposit3 := types.NewDeposit(proposal3.ProposalID, TestAddrs[1], oneCoins)
	_, err = keeper.AddDeposit(ctx, deposit3.ProposalID, deposit3.Depositor, deposit3.Amount)
	require.NoError(t, err)

	proposal3.TotalDeposit = proposal3.TotalDeposit.Add(deposit3.Amount...)

	// TestAddrs[1] deposits on proposals #2 & #3
	deposit4 := types.NewDeposit(proposal2.ProposalID, TestAddrs[1], depositParams.MinDeposit)
	_, err = keeper.AddDeposit(ctx, deposit4.ProposalID, deposit4.Depositor, deposit4.Amount)
	require.NoError(t, err)

	proposal2.TotalDeposit = proposal2.TotalDeposit.Add(deposit4.Amount...)
	proposal2.Status = types.StatusVotingPeriod
	proposal2.VotingEndTime = proposal2.VotingEndTime.Add(types.DefaultPeriod)

	deposit5 := types.NewDeposit(proposal3.ProposalID, TestAddrs[1], depositParams.MinDeposit)
	_, err = keeper.AddDeposit(ctx, deposit5.ProposalID, deposit5.Depositor, deposit5.Amount)
	require.NoError(t, err)

	proposal3.TotalDeposit = proposal3.TotalDeposit.Add(deposit5.Amount...)
	proposal3.Status = types.StatusVotingPeriod
	proposal3.VotingEndTime = proposal3.VotingEndTime.Add(types.DefaultPeriod)
	// total deposit of TestAddrs[1] on proposal #3 is worth the proposal deposit + individual deposit
	deposit5.Amount = deposit5.Amount.Add(deposit3.Amount...)

	// check deposits on proposal1 match individual deposits
	deposits := getQueriedDeposits(t, ctx, keeper.cdc, querier, proposal1.ProposalID)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit1, deposits[0])

	deposit := getQueriedDeposit(t, ctx, keeper.cdc, querier, proposal1.ProposalID, TestAddrs[0])
	require.Equal(t, deposit1, deposit)

	// check deposits on proposal2 match individual deposits
	deposits = getQueriedDeposits(t, ctx, keeper.cdc, querier, proposal2.ProposalID)
	require.Len(t, deposits, 2)
	// NOTE order of deposits is determined by the addresses
	require.Equal(t, deposit2, deposits[0])
	require.Equal(t, deposit4, deposits[1])

	// check deposits on proposal3 match individual deposits
	deposits = getQueriedDeposits(t, ctx, keeper.cdc, querier, proposal3.ProposalID)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit5, deposits[0])

	deposit = getQueriedDeposit(t, ctx, keeper.cdc, querier, proposal3.ProposalID, TestAddrs[1])
	require.Equal(t, deposit5, deposit)

	// Only proposal #1 should be in types.Deposit Period
	proposals := getQueriedProposals(t, ctx, keeper.cdc, querier, nil, nil, types.StatusDepositPeriod, 1, 0)
	require.Len(t, proposals, 1)
	require.Equal(t, proposal1, proposals[0])

	// Only proposals #2 and #3 should be in Voting Period
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, nil, nil, types.StatusVotingPeriod, 1, 0)
	require.Len(t, proposals, 2)
	require.Equal(t, proposal2, proposals[0])
	require.Equal(t, proposal3, proposals[1])

	// Addrs[0] votes on proposals #2 & #3
	vote1 := types.NewVote(proposal2.ProposalID, TestAddrs[0], types.OptionYes)
	vote2 := types.NewVote(proposal3.ProposalID, TestAddrs[0], types.OptionYes)
	keeper.SetVote(ctx, vote1)
	keeper.SetVote(ctx, vote2)

	// Addrs[1] votes on proposal #3
	vote3 := types.NewVote(proposal3.ProposalID, TestAddrs[1], types.OptionYes)
	keeper.SetVote(ctx, vote3)

	// Test query voted by TestAddrs[0]
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, nil, TestAddrs[0], types.StatusNil, 1, 0)
	require.Equal(t, proposal2, proposals[0])
	require.Equal(t, proposal3, proposals[1])

	// Test query votes on types.Proposal 2
	votes := getQueriedVotes(t, ctx, keeper.cdc, querier, proposal2.ProposalID, 1, 0)
	require.Len(t, votes, 1)
	require.Equal(t, vote1, votes[0])

	vote := getQueriedVote(t, ctx, keeper.cdc, querier, proposal2.ProposalID, TestAddrs[0])
	require.Equal(t, vote1, vote)

	// Test query votes on types.Proposal 3
	votes = getQueriedVotes(t, ctx, keeper.cdc, querier, proposal3.ProposalID, 1, 0)
	require.Len(t, votes, 2)
	require.Equal(t, vote2, votes[0])
	require.Equal(t, vote3, votes[1])

	// Test query all proposals
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, nil, nil, types.StatusNil, 1, 0)
	require.Equal(t, proposal1, proposals[0])
	require.Equal(t, proposal2, proposals[1])
	require.Equal(t, proposal3, proposals[2])

	// Test query voted by TestAddrs[1]
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, nil, TestAddrs[1], types.StatusNil, 1, 0)
	require.Equal(t, proposal3.ProposalID, proposals[0].ProposalID)

	// Test query deposited by TestAddrs[0]
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, TestAddrs[0], nil, types.StatusNil, 1, 0)
	require.Equal(t, proposal1.ProposalID, proposals[0].ProposalID)

	// Test query deposited by addr2
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, TestAddrs[1], nil, types.StatusNil, 1, 0)
	require.Equal(t, proposal2.ProposalID, proposals[0].ProposalID)
	require.Equal(t, proposal3.ProposalID, proposals[1].ProposalID)

	// Test query voted AND deposited by addr1
	proposals = getQueriedProposals(t, ctx, keeper.cdc, querier, TestAddrs[0], TestAddrs[0], types.StatusNil, 1, 0)
	require.Equal(t, proposal2.ProposalID, proposals[0].ProposalID)
}

func TestPaginatedVotesQuery(t *testing.T) {
	ctx, _, keeper, _, _ := createTestInput(t, false, 1000)

	proposal := types.Proposal{
		ProposalID: 100,
		Status:     types.StatusVotingPeriod,
	}
	keeper.SetProposal(ctx, proposal)

	votes := make([]types.Vote, 20)
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	addr := make(sdk.AccAddress, 20)
	for i := range votes {
		rand.Read(addr)
		vote := types.Vote{
			ProposalID: proposal.ProposalID,
			Voter:      addr,
			Option:     types.OptionYes,
		}
		votes[i] = vote
		keeper.SetVote(ctx, vote)
	}

	querier := NewQuerier(keeper)

	// keeper preserves consistent order for each query, but this is not the insertion order
	all := getQueriedVotes(t, ctx, keeper.cdc, querier, proposal.ProposalID, 1, 0)
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
			votes := getQueriedVotes(t, ctx, keeper.cdc, querier, proposal.ProposalID, tc.page, tc.limit)
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}
