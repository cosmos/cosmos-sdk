package gov

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// query endpoints supported by the governance Querier
const (
	QueryProposals = "proposals"
	QueryProposal  = "proposal"
	QueryDeposits  = "deposits"
	QueryDeposit   = "deposit"
	QueryVotes     = "votes"
	QueryVote      = "vote"
	QueryTally     = "tally"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case QueryProposals:
			return queryProposals(ctx, path[1:], req, keeper)
		case QueryProposal:
			return queryProposal(ctx, path[1:], req, keeper)
		case QueryDeposits:
			return queryDeposits(ctx, path[1:], req, keeper)
		case QueryDeposit:
			return queryDeposit(ctx, path[1:], req, keeper)
		case QueryVotes:
			return queryVotes(ctx, path[1:], req, keeper)
		case QueryVote:
			return queryVote(ctx, path[1:], req, keeper)
		case QueryTally:
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

// nolint: unparam
func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryProposalParams
	err2 := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	proposal := keeper.GetProposal(ctx, params.ProposalID)
	if proposal == nil {
		return []byte{}, ErrUnknownProposal(DefaultCodespace, params.ProposalID)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, proposal)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

// Params for query 'custom/gov/deposit'
type QueryDepositParams struct {
	ProposalID int64
	Depositer  sdk.AccAddress
}

// nolint: unparam
func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryDepositParams
	err2 := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositer)
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, deposit)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

// Params for query 'custom/gov/vote'
type QueryVoteParams struct {
	ProposalID int64
	Voter      sdk.AccAddress
}

// nolint: unparam
func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryVoteParams
	err2 := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, vote)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

// Params for query 'custom/gov/deposits'
type QueryDepositsParams struct {
	ProposalID int64
}

// nolint: unparam
func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryDepositParams
	err2 := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	var deposits []Deposit
	depositsIterator := keeper.GetDeposits(ctx, params.ProposalID)
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := Deposit{}
		keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, deposits)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

// Params for query 'custom/gov/votes'
type QueryVotesParams struct {
	ProposalID int64
}

// nolint: unparam
func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryVotesParams
	err2 := keeper.cdc.UnmarshalJSON(req.Data, &params)

	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	var votes []Vote
	votesIterator := keeper.GetVotes(ctx, params.ProposalID)
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
		votes = append(votes, vote)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, votes)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

// Params for query 'custom/gov/proposals'
type QueryProposalsParams struct {
	Voter              sdk.AccAddress
	Depositer          sdk.AccAddress
	ProposalStatus     ProposalStatus
	NumLatestProposals int64
}

// nolint: unparam
func queryProposals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryProposalsParams
	err2 := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err2 != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	proposals := keeper.GetProposalsFiltered(ctx, params.Voter, params.Depositer, params.ProposalStatus, params.NumLatestProposals)

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, proposals)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}

// Params for query 'custom/gov/tally'
type QueryTallyParams struct {
	ProposalID int64
}

// nolint: unparam
func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	// TODO: Dependant on #1914

	var proposalID int64
	err2 := keeper.cdc.UnmarshalJSON(req.Data, proposalID)
	if err2 != nil {
		return res, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", err2.Error()))
	}

	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return res, ErrUnknownProposal(DefaultCodespace, proposalID)
	}

	var tallyResult TallyResult

	if proposal.GetStatus() == StatusDepositPeriod {
		tallyResult = EmptyTallyResult()
	} else if proposal.GetStatus() == StatusPassed || proposal.GetStatus() == StatusRejected {
		tallyResult = proposal.GetTallyResult()
	} else {
		_, tallyResult, _ = tally(ctx, keeper, proposal)
	}

	bz, err2 := codec.MarshalJSONIndent(keeper.cdc, tallyResult)
	if err2 != nil {
		panic("could not marshal result to JSON")
	}
	return bz, nil
}
