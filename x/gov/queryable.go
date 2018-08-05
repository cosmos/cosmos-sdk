package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case "proposal":
			return queryProposal(ctx, path[1:], req, keeper)
		case "deposit":
			return queryDeposit(ctx, path[1:], req, keeper)
		case "vote":
			return queryVote(ctx, path[1:], req, keeper)
		case "deposits":
			return queryDeposits(ctx, path[1:], req, keeper)
		case "votes":
			return queryVotes(ctx, path[1:], req, keeper)
		case "proposals":
			return queryProposals(ctx, path[1:], req, keeper)
		case "tally":
			return queryTally(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown gov query endpoint")
		}
	}
}

// Params for query 'custom/gov/proposal'
type QueryProposalParams struct {
	ProposalID int64
}

func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryProposalParams
	err2 := keeper.cdc.UnmarshalBinary(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	proposal := keeper.GetProposal(ctx, params.ProposalID)
	if proposal == nil {
		return []byte{}, ErrUnknownProposal(DefaultCodespace, params.ProposalID)
	}
	return keeper.cdc.MustMarshalBinary(proposal), nil
}

// Params for query 'custom/gov/deposit'
type QueryDepositParams struct {
	ProposalID int64
	Depositer  sdk.AccAddress
}

func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryDepositParams
	err2 := keeper.cdc.UnmarshalBinary(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositer)
	return keeper.cdc.MustMarshalBinary(deposit), nil
}

// Params for query 'custom/gov/vote'
type QueryVoteParams struct {
	ProposalID int64
	Voter      sdk.AccAddress
}

func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryVoteParams
	err2 := keeper.cdc.UnmarshalBinary(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	return keeper.cdc.MustMarshalBinary(vote), nil
}

// Params for query 'custom/gov/deposits'
type QueryDepositsParams struct {
	ProposalID int64
}

func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryDepositParams
	err2 := keeper.cdc.UnmarshalBinary(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	var deposits []Deposit
	depositsIterator := keeper.GetDeposits(ctx, params.ProposalID)
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := Deposit{}
		keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	return keeper.cdc.MustMarshalBinary(deposits), nil
}

// Params for query 'custom/gov/votes'
type QueryVotesParams struct {
	ProposalID int64
}

func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryVotesParams
	err2 := keeper.cdc.UnmarshalBinary(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	var votes []Vote
	votesIterator := keeper.GetVotes(ctx, params.ProposalID)
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
		votes = append(votes, vote)
	}

	return keeper.cdc.MustMarshalBinary(votes), nil
}

// Params for query 'custom/gov/proposals'
type QueryProposalsParams struct {
	Voter              sdk.AccAddress
	Depositer          sdk.AccAddress
	ProposalStatus     ProposalStatus
	NumLatestProposals int64
}

func queryProposals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryProposalsParams
	err2 := keeper.cdc.UnmarshalBinary(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest("incorrectly formatted request data")
	}

	proposals := keeper.GetProposalsFiltered(ctx, params.Voter, params.Depositer, params.ProposalStatus, params.NumLatestProposals)

	bz := keeper.cdc.MustMarshalBinary(proposals)
	return bz, nil
}

// Params for query 'custom/gov/tally'
type QueryTallyParams struct {
	ProposalID int64
}

func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	// TODO: Dependant on #1914

	// var proposalID int64
	// err2 := keeper.cdc.UnmarshalBinary(req.Data, proposalID)
	// if err2 != nil {
	// 	return []byte{}, sdk.ErrUnknownRequest()
	// }

	// proposal := keeper.GetProposal(ctx, proposalID)
	// if proposal == nil {
	// 	return []byte{}, ErrUnknownProposal(DefaultCodespace, proposalID)
	// }
	// _, tallyResult, _ := tally(ctx, keeper, proposal)
	return nil, nil
}
