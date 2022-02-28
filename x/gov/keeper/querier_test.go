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
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier) (v1beta2.DepositParams, v1beta2.VotingParams, v1beta2.TallyParams) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryParams, v1beta2.ParamDeposit}, "/"),
		Data: []byte{},
	}

	bz, err := querier(ctx, []string{v1beta2.QueryParams, v1beta2.ParamDeposit}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var depositParams v1beta2.DepositParams
	require.NoError(t, cdc.UnmarshalJSON(bz, &depositParams))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryParams, v1beta2.ParamVoting}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{v1beta2.QueryParams, v1beta2.ParamVoting}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var votingParams v1beta2.VotingParams
	require.NoError(t, cdc.UnmarshalJSON(bz, &votingParams))

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryParams, v1beta2.ParamTallying}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{v1beta2.QueryParams, v1beta2.ParamTallying}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var tallyParams v1beta2.TallyParams
	require.NoError(t, cdc.UnmarshalJSON(bz, &tallyParams))

	return depositParams, votingParams, tallyParams
}

func getQueriedProposals(
	t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier,
	depositor, voter sdk.AccAddress, status v1beta2.ProposalStatus, page, limit int,
) []*v1beta2.Proposal {

	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryProposals}, "/"),
		Data: cdc.MustMarshalJSON(v1beta2.NewQueryProposalsParams(page, limit, status, voter, depositor)),
	}

	bz, err := querier(ctx, []string{v1beta2.QueryProposals}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var proposals v1beta2.Proposals
	require.NoError(t, cdc.UnmarshalJSON(bz, &proposals))

	return proposals
}

func getQueriedDeposit(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, proposalID uint64, depositor sdk.AccAddress) v1beta2.Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryDeposit}, "/"),
		Data: cdc.MustMarshalJSON(v1beta2.NewQueryDepositParams(proposalID, depositor)),
	}

	bz, err := querier(ctx, []string{v1beta2.QueryDeposit}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var deposit v1beta2.Deposit
	require.NoError(t, cdc.UnmarshalJSON(bz, &deposit))

	return deposit
}

func getQueriedDeposits(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, proposalID uint64) []v1beta2.Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryDeposits}, "/"),
		Data: cdc.MustMarshalJSON(v1beta2.NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{v1beta2.QueryDeposits}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var deposits []v1beta2.Deposit
	require.NoError(t, cdc.UnmarshalJSON(bz, &deposits))

	return deposits
}

func getQueriedVote(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier, proposalID uint64, voter sdk.AccAddress) v1beta2.Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(v1beta2.NewQueryVoteParams(proposalID, voter)),
	}

	bz, err := querier(ctx, []string{v1beta2.QueryVote}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var vote v1beta2.Vote
	require.NoError(t, cdc.UnmarshalJSON(bz, &vote))

	return vote
}

func getQueriedVotes(t *testing.T, ctx sdk.Context, cdc *codec.LegacyAmino, querier sdk.Querier,
	proposalID uint64, page, limit int) []v1beta2.Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, v1beta2.QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(v1beta2.NewQueryProposalVotesParams(proposalID, page, limit)),
	}

	bz, err := querier(ctx, []string{v1beta2.QueryVotes}, query)
	require.NoError(t, err)
	require.NotNil(t, bz)

	var votes []v1beta2.Vote
	require.NoError(t, cdc.UnmarshalJSON(bz, &votes))

	return votes
}

func TestQueries(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	legacyQuerierCdc := app.LegacyAmino()
	querier := keeper.NewQuerier(app.GovKeeper, legacyQuerierCdc)

	TestAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(20000001))

	oneCoins := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1))
	consCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 10)))

	tp := TestProposal

	depositParams, _, _ := getQueriedParams(t, ctx, legacyQuerierCdc, querier)

	// TestAddrs[0] proposes (and deposits) proposals #1 and #2
	proposal1, err := app.GovKeeper.SubmitProposal(ctx, tp, nil)
	require.NoError(t, err)
	deposit1 := v1beta2.NewDeposit(proposal1.Id, TestAddrs[0], oneCoins)
	depositer1, err := sdk.AccAddressFromBech32(deposit1.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit1.ProposalId, depositer1, deposit1.Amount)
	require.NoError(t, err)

	proposal1.TotalDeposit = sdk.NewCoins(proposal1.TotalDeposit...).Add(deposit1.Amount...)

	proposal2, err := app.GovKeeper.SubmitProposal(ctx, tp, nil)
	require.NoError(t, err)
	deposit2 := v1beta2.NewDeposit(proposal2.Id, TestAddrs[0], consCoins)
	depositer2, err := sdk.AccAddressFromBech32(deposit2.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit2.ProposalId, depositer2, deposit2.Amount)
	require.NoError(t, err)

	proposal2.TotalDeposit = sdk.NewCoins(proposal2.TotalDeposit...).Add(deposit2.Amount...)

	// TestAddrs[1] proposes (and deposits) on proposal #3
	proposal3, err := app.GovKeeper.SubmitProposal(ctx, tp, nil)
	require.NoError(t, err)
	deposit3 := v1beta2.NewDeposit(proposal3.Id, TestAddrs[1], oneCoins)
	depositer3, err := sdk.AccAddressFromBech32(deposit3.Depositor)
	require.NoError(t, err)

	_, err = app.GovKeeper.AddDeposit(ctx, deposit3.ProposalId, depositer3, deposit3.Amount)
	require.NoError(t, err)

	proposal3.TotalDeposit = sdk.NewCoins(proposal3.TotalDeposit...).Add(deposit3.Amount...)

	// TestAddrs[1] deposits on proposals #2 & #3
	deposit4 := v1beta2.NewDeposit(proposal2.Id, TestAddrs[1], depositParams.MinDeposit)
	depositer4, err := sdk.AccAddressFromBech32(deposit4.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit4.ProposalId, depositer4, deposit4.Amount)
	require.NoError(t, err)

	proposal2.TotalDeposit = sdk.NewCoins(proposal2.TotalDeposit...).Add(deposit4.Amount...)
	proposal2.Status = v1beta2.StatusVotingPeriod
	votingEndTime := ctx.BlockTime().Add(v1beta2.DefaultPeriod)
	proposal2.VotingEndTime = &votingEndTime

	deposit5 := v1beta2.NewDeposit(proposal3.Id, TestAddrs[1], depositParams.MinDeposit)
	depositer5, err := sdk.AccAddressFromBech32(deposit5.Depositor)
	require.NoError(t, err)
	_, err = app.GovKeeper.AddDeposit(ctx, deposit5.ProposalId, depositer5, deposit5.Amount)
	require.NoError(t, err)

	proposal3.TotalDeposit = sdk.NewCoins(proposal3.TotalDeposit...).Add(deposit5.Amount...)
	proposal3.Status = v1beta2.StatusVotingPeriod
	votingEndTime = ctx.BlockTime().Add(v1beta2.DefaultPeriod)
	proposal3.VotingEndTime = &votingEndTime
	// total deposit of TestAddrs[1] on proposal #3 is worth the proposal deposit + individual deposit
	deposit5.Amount = sdk.NewCoins(deposit5.Amount...).Add(deposit3.Amount...)

	// check deposits on proposal1 match individual deposits

	deposits := getQueriedDeposits(t, ctx, legacyQuerierCdc, querier, proposal1.Id)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit1, deposits[0])

	deposit := getQueriedDeposit(t, ctx, legacyQuerierCdc, querier, proposal1.Id, TestAddrs[0])
	require.Equal(t, deposit1, deposit)

	// check deposits on proposal2 match individual deposits
	deposits = getQueriedDeposits(t, ctx, legacyQuerierCdc, querier, proposal2.Id)
	require.Len(t, deposits, 2)
	// NOTE order of deposits is determined by the addresses
	require.Equal(t, deposit2, deposits[0])
	require.Equal(t, deposit4, deposits[1])

	// check deposits on proposal3 match individual deposits
	deposits = getQueriedDeposits(t, ctx, legacyQuerierCdc, querier, proposal3.Id)
	require.Len(t, deposits, 1)
	require.Equal(t, deposit5, deposits[0])

	deposit = getQueriedDeposit(t, ctx, legacyQuerierCdc, querier, proposal3.Id, TestAddrs[1])
	require.Equal(t, deposit5, deposit)

	// Only proposal #1 should be in v1beta2.Deposit Period
	proposals := getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, nil, v1beta2.StatusDepositPeriod, 1, 0)
	require.Len(t, proposals, 1)
	require.Equal(t, proposal1, *proposals[0])

	// Only proposals #2 and #3 should be in Voting Period
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, nil, v1beta2.StatusVotingPeriod, 1, 0)
	require.Len(t, proposals, 2)
	checkEqualProposal(t, proposal2, *proposals[0])
	checkEqualProposal(t, proposal3, *proposals[1])

	// Addrs[0] votes on proposals #2 & #3
	vote1 := v1beta2.NewVote(proposal2.Id, TestAddrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes))
	vote2 := v1beta2.NewVote(proposal3.Id, TestAddrs[0], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes))
	app.GovKeeper.SetVote(ctx, vote1)
	app.GovKeeper.SetVote(ctx, vote2)

	// Addrs[1] votes on proposal #3
	vote3 := v1beta2.NewVote(proposal3.Id, TestAddrs[1], v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes))
	app.GovKeeper.SetVote(ctx, vote3)

	// Test query voted by TestAddrs[0]
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, TestAddrs[0], v1beta2.StatusNil, 1, 0)
	checkEqualProposal(t, proposal2, *proposals[0])
	checkEqualProposal(t, proposal3, *proposals[1])

	// Test query votes on v1beta2.Proposal 2
	votes := getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal2.Id, 1, 0)
	require.Len(t, votes, 1)
	checkEqualVotes(t, vote1, votes[0])

	vote := getQueriedVote(t, ctx, legacyQuerierCdc, querier, proposal2.Id, TestAddrs[0])
	checkEqualVotes(t, vote1, vote)

	// Test query votes on v1beta2.Proposal 3
	votes = getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal3.Id, 1, 0)
	require.Len(t, votes, 2)
	checkEqualVotes(t, vote2, votes[0])
	checkEqualVotes(t, vote3, votes[1])

	// Test query all proposals
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, nil, v1beta2.StatusNil, 1, 0)
	checkEqualProposal(t, proposal1, *proposals[0])
	checkEqualProposal(t, proposal2, *proposals[1])
	checkEqualProposal(t, proposal3, *proposals[2])

	// Test query voted by TestAddrs[1]
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, nil, TestAddrs[1], v1beta2.StatusNil, 1, 0)
	require.Equal(t, proposal3.Id, proposals[0].Id)

	// Test query deposited by TestAddrs[0]
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, TestAddrs[0], nil, v1beta2.StatusNil, 1, 0)
	require.Equal(t, proposal1.Id, proposals[0].Id)

	// Test query deposited by addr2
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, TestAddrs[1], nil, v1beta2.StatusNil, 1, 0)
	require.Equal(t, proposal2.Id, proposals[0].Id)
	require.Equal(t, proposal3.Id, proposals[1].Id)

	// Test query voted AND deposited by addr1
	proposals = getQueriedProposals(t, ctx, legacyQuerierCdc, querier, TestAddrs[0], TestAddrs[0], v1beta2.StatusNil, 1, 0)
	require.Equal(t, proposal2.Id, proposals[0].Id)
}

func TestPaginatedVotesQuery(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	legacyQuerierCdc := app.LegacyAmino()

	proposal := v1beta2.Proposal{
		Id:     100,
		Status: v1beta2.StatusVotingPeriod,
	}

	app.GovKeeper.SetProposal(ctx, proposal)

	votes := make([]v1beta2.Vote, 20)
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
		vote := v1beta2.Vote{
			ProposalId: proposal.Id,
			Voter:      genAddr(),
			Options:    v1beta2.NewNonSplitVoteOption(v1beta2.OptionYes),
		}
		votes[i] = vote
		app.GovKeeper.SetVote(ctx, vote)
	}

	querier := keeper.NewQuerier(app.GovKeeper, legacyQuerierCdc)

	// keeper preserves consistent order for each query, but this is not the insertion order
	all := getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal.Id, 1, 0)
	require.Equal(t, len(all), len(votes))

	type testCase struct {
		description string
		page        int
		limit       int
		votes       []v1beta2.Vote
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
			votes := getQueriedVotes(t, ctx, legacyQuerierCdc, querier, proposal.Id, tc.page, tc.limit)
			require.Equal(t, len(tc.votes), len(votes))
			for i := range votes {
				require.Equal(t, tc.votes[i], votes[i])
			}
		})
	}
}

// checkEqualVotes checks that two votes are equal, without taking into account
// graceful fallback for `Option`.
// When querying, the keeper populates the `vote.Option` field when there's
// only 1 vote, this function checks equality of structs while skipping that
// field.
func checkEqualVotes(t *testing.T, vote1, vote2 v1beta2.Vote) {
	require.Equal(t, vote1.Options, vote2.Options)
	require.Equal(t, vote1.Voter, vote2.Voter)
	require.Equal(t, vote1.ProposalId, vote2.ProposalId)
}

// checkEqualProposal checks that 2 proposals are equal.
// When decoding with Amino, there are weird cases where the voting times
// are actually equal, but `require.Equal()` says they are not:
//
// Diff:
// --- Expected
// +++ Actual
// @@ -68,3 +68,7 @@
// 	},
// - VotingStartTime: (*time.Time)(<nil>),
// + VotingStartTime: (*time.Time)({
// +  wall: (uint64) 0,
// +  ext: (int64) 0,
// +  loc: (*time.Location)(<nil>)
// + }),
func checkEqualProposal(t *testing.T, p1, p2 v1beta2.Proposal) {
	require.Equal(t, p1.Id, p2.Id)
	require.Equal(t, p1.Messages, p2.Messages)
	require.Equal(t, p1.Status, p2.Status)
	require.Equal(t, p1.FinalTallyResult, p2.FinalTallyResult)
	require.Equal(t, p1.SubmitTime, p2.SubmitTime)
	require.Equal(t, p1.DepositEndTime, p2.DepositEndTime)
	require.Equal(t, p1.TotalDeposit, p2.TotalDeposit)
	require.Equal(t, convertNilToDefault(p1.VotingStartTime), convertNilToDefault(p2.VotingStartTime))
	require.Equal(t, convertNilToDefault(p1.VotingEndTime), convertNilToDefault(p2.VotingEndTime))
}

// convertNilToDefault converts a (*time.Time)(<nil>) into a (*time.Time)(<default>)).
// In proto structs dealing with time, we use *time.Time, which can be nil.
// However, when using Amino, a nil time is unmarshalled into
// (*time.Time)(<default>)), which is Jan 1st 1970.
// This function converts a nil time to a default time, to check that they are
// actually equal.
func convertNilToDefault(t *time.Time) *time.Time {
	if t == nil {
		return &time.Time{}
	}

	return t
}
