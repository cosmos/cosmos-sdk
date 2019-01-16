package gov

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const custom = "custom"

func getQueriedParams(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier) (DepositParams, VotingParams, TallyParams) {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryParams, ParamDeposit}, "/"),
		Data: []byte{},
	}

	bz, err := querier(ctx, []string{QueryParams, ParamDeposit}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var depositParams DepositParams
	err2 := cdc.UnmarshalJSON(bz, &depositParams)
	require.Nil(t, err2)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryParams, ParamVoting}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{QueryParams, ParamVoting}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var votingParams VotingParams
	err2 = cdc.UnmarshalJSON(bz, &votingParams)
	require.Nil(t, err2)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryParams, ParamTallying}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{QueryParams, ParamTallying}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var tallyParams TallyParams
	err2 = cdc.UnmarshalJSON(bz, &tallyParams)
	require.Nil(t, err2)

	return depositParams, votingParams, tallyParams
}

func getQueriedProposal(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) Proposal {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryProposal}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{QueryProposal}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var proposal Proposal
	err2 := cdc.UnmarshalJSON(bz, proposal)
	require.Nil(t, err2)
	return proposal
}

func getQueriedProposals(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, depositor, voter sdk.AccAddress, status ProposalStatus, limit uint64) []Proposal {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryProposals}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalsParams(status, limit, voter, depositor)),
	}

	bz, err := querier(ctx, []string{QueryProposal}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var proposals []Proposal
	err2 := cdc.UnmarshalJSON(bz, proposals)
	require.Nil(t, err2)
	return proposals
}

func getQueriedDeposit(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64, depositor sdk.AccAddress) Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryDeposit}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryDepositParams(proposalID, depositor)),
	}

	bz, err := querier(ctx, []string{QueryDeposits}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var deposit Deposit
	err2 := cdc.UnmarshalJSON(bz, deposit)
	require.Nil(t, err2)
	return deposit
}

func getQueriedDeposits(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) []Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryDeposits}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{QueryDeposits}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var deposits []Deposit
	err2 := cdc.UnmarshalJSON(bz, &deposits)
	require.Nil(t, err2)
	return deposits
}

func getQueriedVote(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64, voter sdk.AccAddress) Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryVoteParams(proposalID, voter)),
	}

	bz, err := querier(ctx, []string{QueryVote}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var vote Vote
	err2 := cdc.UnmarshalJSON(bz, &vote)
	require.Nil(t, err2)
	return vote
}

func getQueriedVotes(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) []Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{QueryVotes}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var votes []Vote
	err2 := cdc.UnmarshalJSON(bz, &votes)
	require.Nil(t, err2)
	return votes
}

func getQueriedTally(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) TallyResult {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, QuerierRoute, QueryTally}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{QueryTally}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var tally TallyResult
	err2 := cdc.UnmarshalJSON(bz, &tally)
	require.Nil(t, err2)
	return tally
}

func testQueryParams(t *testing.T) {
	cdc := codec.New()
	mapp, keeper, _, _, _, _ := getMockApp(t, 1000, GenesisState{}, nil)
	querier := NewQuerier(keeper)
	ctx := mapp.NewContext(false, abci.Header{})

	getQueriedParams(t, ctx, cdc, querier)
}

func testQueries(t *testing.T) {
	cdc := codec.New()
	mapp, keeper, _, addrs, _, _ := getMockApp(t, 1000, GenesisState{}, nil)
	querier := NewQuerier(keeper)
	handler := NewHandler(keeper)
	ctx := mapp.NewContext(false, abci.Header{})

	depositParams, _, _ := getQueriedParams(t, ctx, cdc, querier)

	// addrs[0] proposes (and deposits) proposals #1 and #2
	res := handler(ctx, NewMsgSubmitProposal("title", "description", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("dummycoin", 1)}))
	var proposalID1 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID1)

	res = handler(ctx, NewMsgSubmitProposal("title", "description", ProposalTypeText, addrs[0], sdk.Coins{sdk.NewInt64Coin("dummycoin", 1)}))
	var proposalID2 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID2)

	// addrs[1] proposes (and deposits) proposals #3
	res = handler(ctx, NewMsgSubmitProposal("title", "description", ProposalTypeText, addrs[1], sdk.Coins{sdk.NewInt64Coin("dummycoin", 1)}))
	var proposalID3 uint64
	cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID3)

	// addrs[1] deposits on proposals #2 & #3
	res = handler(ctx, NewMsgDeposit(addrs[1], proposalID2, depositParams.MinDeposit))
	res = handler(ctx, NewMsgDeposit(addrs[1], proposalID3, depositParams.MinDeposit))

	// check deposits on proposal1 match individual deposits
	deposits := getQueriedDeposits(t, ctx, cdc, querier, proposalID1)
	require.Len(t, deposits, 1)
	deposit := getQueriedDeposit(t, ctx, cdc, querier, proposalID1, addrs[0])
	require.Equal(t, deposit, deposits[0])

	// check deposits on proposal2 match individual deposits
	deposits = getQueriedDeposits(t, ctx, cdc, querier, proposalID2)
	require.Len(t, deposits, 2)
	deposit = getQueriedDeposit(t, ctx, cdc, querier, proposalID2, addrs[0])
	require.True(t, deposit.Equals(deposits[0]))
	deposit = getQueriedDeposit(t, ctx, cdc, querier, proposalID2, addrs[1])
	require.True(t, deposit.Equals(deposits[1]))

	// check deposits on proposal3 match individual deposits
	deposits = getQueriedDeposits(t, ctx, cdc, querier, proposalID3)
	require.Len(t, deposits, 1)
	deposit = getQueriedDeposit(t, ctx, cdc, querier, proposalID3, addrs[1])
	require.Equal(t, deposit, deposits[0])

	// Only proposal #1 should be in Deposit Period
	proposals := getQueriedProposals(t, ctx, cdc, querier, nil, nil, StatusDepositPeriod, 0)
	require.Len(t, proposals, 1)
	require.Equal(t, proposalID1, proposals[0].GetProposalID())
	// Only proposals #2 and #3 should be in Voting Period
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, nil, StatusVotingPeriod, 0)
	require.Len(t, proposals, 2)
	require.Equal(t, proposalID2, proposals[0].GetProposalID())
	require.Equal(t, proposalID3, proposals[1].GetProposalID())

	// Addrs[0] votes on proposals #2 & #3
	handler(ctx, NewMsgVote(addrs[0], proposalID2, OptionYes))
	handler(ctx, NewMsgVote(addrs[0], proposalID3, OptionYes))

	// Addrs[1] votes on proposal #3
	handler(ctx, NewMsgVote(addrs[1], proposalID3, OptionYes))

	// Test query voted by addrs[0]
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, addrs[0], StatusNil, 0)
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[1]).GetProposalID())

	// Test query votes on Proposal 2
	votes := getQueriedVotes(t, ctx, cdc, querier, proposalID2)
	require.Len(t, votes, 1)
	require.Equal(t, addrs[0], votes[0].Voter)
	vote := getQueriedVote(t, ctx, cdc, querier, proposalID2, addrs[0])
	require.Equal(t, vote, votes[0])

	// Test query votes on Proposal 3
	votes = getQueriedVotes(t, ctx, cdc, querier, proposalID3)
	require.Len(t, votes, 2)
	require.True(t, addrs[0].String() == votes[0].Voter.String())
	require.True(t, addrs[1].String() == votes[0].Voter.String())

	// Test proposals queries with filters

	// Test query all proposals
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, nil, StatusNil, 0)
	require.Equal(t, proposalID1, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID2, (proposals[1]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[2]).GetProposalID())

	// Test query voted by addrs[1]
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, addrs[1], StatusNil, 0)
	require.Equal(t, proposalID3, (proposals[0]).GetProposalID())

	// Test query deposited by addrs[0]
	proposals = getQueriedProposals(t, ctx, cdc, querier, addrs[0], nil, StatusNil, 0)
	require.Equal(t, proposalID1, (proposals[0]).GetProposalID())

	// Test query deposited by addr2
	proposals = getQueriedProposals(t, ctx, cdc, querier, addrs[1], nil, StatusNil, 0)
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())
	require.Equal(t, proposalID3, (proposals[1]).GetProposalID())

	// Test query voted AND deposited by addr1
	proposals = getQueriedProposals(t, ctx, cdc, querier, addrs[0], addrs[0], StatusNil, 0)
	require.Equal(t, proposalID2, (proposals[0]).GetProposalID())

	// Test Tally Query
	tally := getQueriedTally(t, ctx, cdc, querier, proposalID2)
	require.True(t, !tally.Equals(EmptyTallyResult()))
}
