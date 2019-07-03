package keeper

import (
	"strings"
	"testing"

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
	require.Nil(t, err)
	require.NotNil(t, bz)

	var depositParams types.DepositParams
	err2 := cdc.UnmarshalJSON(bz, &depositParams)
	require.Nil(t, err2)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryParams, types.ParamVoting}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{types.QueryParams, types.ParamVoting}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var votingParams types.VotingParams
	err2 = cdc.UnmarshalJSON(bz, &votingParams)
	require.Nil(t, err2)

	query = abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryParams, types.ParamTallying}, "/"),
		Data: []byte{},
	}

	bz, err = querier(ctx, []string{types.QueryParams, types.ParamTallying}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var tallyParams types.TallyParams
	err2 = cdc.UnmarshalJSON(bz, &tallyParams)
	require.Nil(t, err2)

	return depositParams, votingParams, tallyParams
}

func getQueriedProposal(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) types.Proposal {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposal}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{types.QueryProposal}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var proposal types.Proposal
	err2 := cdc.UnmarshalJSON(bz, proposal)
	require.Nil(t, err2)
	return proposal
}

func getQueriedProposals(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, depositor, voter sdk.AccAddress, status types.ProposalStatus, limit uint64) []types.Proposal {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryProposals}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalsParams(status, limit, voter, depositor)),
	}

	bz, err := querier(ctx, []string{types.QueryProposals}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var proposals Proposals
	err2 := cdc.UnmarshalJSON(bz, &proposals)
	require.Nil(t, err2)
	return proposals
}

func getQueriedDeposit(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64, depositor sdk.AccAddress) types.Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryDeposit}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryDepositParams(proposalID, depositor)),
	}

	bz, err := querier(ctx, []string{types.QueryDeposit}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var deposit types.Deposit
	err2 := cdc.UnmarshalJSON(bz, &deposit)
	require.Nil(t, err2)
	return deposit
}

func getQueriedDeposits(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) []types.Deposit {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.types.QueryDeposits}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{types.types.QueryDeposits}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var deposits []types.Deposit
	err2 := cdc.UnmarshalJSON(bz, &deposits)
	require.Nil(t, err2)
	return deposits
}

func getQueriedVote(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64, voter sdk.AccAddress) types.Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryVoteParams(proposalID, voter)),
	}

	bz, err := querier(ctx, []string{types.QueryVote}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var vote types.Vote
	err2 := cdc.UnmarshalJSON(bz, &vote)
	require.Nil(t, err2)
	return vote
}

func getQueriedVotes(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) []types.Vote {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryVote}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{QueryVotes}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var votes []types.Vote
	err2 := cdc.UnmarshalJSON(bz, &votes)
	require.Nil(t, err2)
	return votes
}

func getQueriedTally(t *testing.T, ctx sdk.Context, cdc *codec.Codec, querier sdk.Querier, proposalID uint64) TallyResult {
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryTally}, "/"),
		Data: cdc.MustMarshalJSON(NewQueryProposalParams(proposalID)),
	}

	bz, err := querier(ctx, []string{types.QueryTally}, query)
	require.Nil(t, err)
	require.NotNil(t, bz)

	var tally TallyResult
	err2 := cdc.UnmarshalJSON(bz, &tally)
	require.Nil(t, err2)
	return tally
}

func TestQueryParams(t *testing.T) {
	cdc := codec.New()
	input := getMockApp(t, 1000, types.GenesisState{}, nil)
	querier := NewQuerier(input.keeper)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.NewContext(false, abci.Header{})

	getQueriedParams(t, ctx, cdc, querier)
}

func TestQueries(t *testing.T) {
	cdc := codec.New()
	input := getMockApp(t, 1000, types.GenesisState{}, nil)
	querier := NewQuerier(input.keeper)
	handler := NewHandler(input.keeper)

	types.RegisterCodec(cdc)

	header := abci.Header{Height: input.mApp.LastBlockHeight() + 1}
	input.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})

	ctx := input.mApp.NewContext(false, abci.Header{})

	depositParams, _, _ := getQueriedParams(t, ctx, cdc, querier)

	// input.addrs[0] proposes (and deposits) proposals #1 and #2
	res := handler(ctx, NewMsgSubmitProposal(testProposal(), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)}, input.addrs[0]))
	var proposalID1 uint64
	require.True(t, res.IsOK())
	cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID1)

	res = handler(ctx, NewMsgSubmitProposal(testProposal(), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10000000)}, input.addrs[0]))
	var proposalID2 uint64
	require.True(t, res.IsOK())
	cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID2)

	// input.addrs[1] proposes (and deposits) proposals #3
	res = handler(ctx, NewMsgSubmitProposal(testProposal(), sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)}, input.addrs[1]))
	var proposalID3 uint64
	require.True(t, res.IsOK())
	cdc.MustUnmarshalBinaryLengthPrefixed(res.Data, &proposalID3)

	// input.addrs[1] deposits on proposals #2 & #3
	res = handler(ctx, NewMsgDeposit(input.addrs[1], proposalID2, depositParams.MinDeposit))
	res = handler(ctx, NewMsgDeposit(input.addrs[1], proposalID3, depositParams.MinDeposit))

	// check deposits on proposal1 match individual deposits
	deposits := getQueriedDeposits(t, ctx, cdc, querier, proposalID1)
	require.Len(t, deposits, 1)
	deposit := getQueriedDeposit(t, ctx, cdc, querier, proposalID1, input.addrs[0])
	require.Equal(t, deposit, deposits[0])

	// check deposits on proposal2 match individual deposits
	deposits = getQueriedDeposits(t, ctx, cdc, querier, proposalID2)
	require.Len(t, deposits, 2)
	deposit = getQueriedDeposit(t, ctx, cdc, querier, proposalID2, input.addrs[0])
	require.True(t, deposit.Equals(deposits[0]))
	deposit = getQueriedDeposit(t, ctx, cdc, querier, proposalID2, input.addrs[1])
	require.True(t, deposit.Equals(deposits[1]))

	// check deposits on proposal3 match individual deposits
	deposits = getQueriedDeposits(t, ctx, cdc, querier, proposalID3)
	require.Len(t, deposits, 1)
	deposit = getQueriedDeposit(t, ctx, cdc, querier, proposalID3, input.addrs[1])
	require.Equal(t, deposit, deposits[0])

	// Only proposal #1 should be in types.Deposit Period
	proposals := getQueriedProposals(t, ctx, cdc, querier, nil, nil, StatusDepositPeriod, 0)
	require.Len(t, proposals, 1)
	require.Equal(t, proposalID1, proposals[0].ProposalID)

	// Only proposals #2 and #3 should be in Voting Period
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, nil, StatusVotingPeriod, 0)
	require.Len(t, proposals, 2)
	require.Equal(t, proposalID2, proposals[0].ProposalID)
	require.Equal(t, proposalID3, proposals[1].ProposalID)

	// Addrs[0] votes on proposals #2 & #3
	require.True(t, handler(ctx, NewMsgVote(input.addrs[0], proposalID2, OptionYes)).IsOK())
	require.True(t, handler(ctx, NewMsgVote(input.addrs[0], proposalID3, OptionYes)).IsOK())

	// Addrs[1] votes on proposal #3
	handler(ctx, NewMsgVote(input.addrs[1], proposalID3, OptionYes))

	// Test query voted by input.addrs[0]
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, input.addrs[0], StatusNil, 0)
	require.Equal(t, proposalID2, (proposals[0]).ProposalID)
	require.Equal(t, proposalID3, (proposals[1]).ProposalID)

	// Test query votes on types.Proposal 2
	votes := getQueriedVotes(t, ctx, cdc, querier, proposalID2)
	require.Len(t, votes, 1)
	require.Equal(t, input.addrs[0], votes[0].Voter)

	vote := getQueriedVote(t, ctx, cdc, querier, proposalID2, input.addrs[0])
	require.Equal(t, vote, votes[0])

	// Test query votes on types.Proposal 3
	votes = getQueriedVotes(t, ctx, cdc, querier, proposalID3)
	require.Len(t, votes, 2)
	require.True(t, input.addrs[0].String() == votes[0].Voter.String())
	require.True(t, input.addrs[1].String() == votes[1].Voter.String())

	// Test proposals queries with filters

	// Test query all proposals
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, nil, StatusNil, 0)
	require.Equal(t, proposalID1, (proposals[0]).ProposalID)
	require.Equal(t, proposalID2, (proposals[1]).ProposalID)
	require.Equal(t, proposalID3, (proposals[2]).ProposalID)

	// Test query voted by input.addrs[1]
	proposals = getQueriedProposals(t, ctx, cdc, querier, nil, input.addrs[1], StatusNil, 0)
	require.Equal(t, proposalID3, (proposals[0]).ProposalID)

	// Test query deposited by input.addrs[0]
	proposals = getQueriedProposals(t, ctx, cdc, querier, input.addrs[0], nil, StatusNil, 0)
	require.Equal(t, proposalID1, (proposals[0]).ProposalID)

	// Test query deposited by addr2
	proposals = getQueriedProposals(t, ctx, cdc, querier, input.addrs[1], nil, StatusNil, 0)
	require.Equal(t, proposalID2, (proposals[0]).ProposalID)
	require.Equal(t, proposalID3, (proposals[1]).ProposalID)

	// Test query voted AND deposited by addr1
	proposals = getQueriedProposals(t, ctx, cdc, querier, input.addrs[0], input.addrs[0], StatusNil, 0)
	require.Equal(t, proposalID2, (proposals[0]).ProposalID)
}
