package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case "proposals":
			return queryProposals(ctx, path[1:], req, keeper)
		case "proposal":
			return queryProposal(ctx, path[1:], req, keeper)
		case "deposits":
			return queryDeposits(ctx, path[1:], req, keeper)
		case "deposit":
			return queryDeposit(ctx, path[1:], req, keeper)
		case "votes":
			return queryVotes(ctx, path[1:], req, keeper)
		case "vote":
			return queryVote(ctx, path[1:], req, keeper)
		case "tally":
			return queryTally(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown gov query endpoint")
		}
	}
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
	errRes := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	proposals := keeper.GetProposalsFiltered(ctx, params.Voter, params.Depositer, params.ProposalStatus, params.NumLatestProposals)

	res, errRes = wire.MarshalJSONIndent(keeper.cdc, proposals)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// Params for query 'custom/gov/proposal'
type QueryProposalParams struct {
	ProposalID int64
}

func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryProposalParams
	errRes := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	proposal := keeper.GetProposal(ctx, params.ProposalID)
	if proposal == nil {
		return []byte{}, ErrUnknownProposal(DefaultCodespace, params.ProposalID)
	}

	res, errRes = wire.MarshalJSONIndent(keeper.cdc, proposal)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// Params for query 'custom/gov/deposit'
type QueryDepositParams struct {
	ProposalID int64
	Depositer  sdk.AccAddress
}

func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryDepositParams
	errRes := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositer)
	res, errRes = wire.MarshalJSONIndent(keeper.cdc, deposit)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// Params for query 'custom/gov/vote'
type QueryVoteParams struct {
	ProposalID int64
	Voter      sdk.AccAddress
}

func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryVoteParams
	errRes := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	res, errRes = wire.MarshalJSONIndent(keeper.cdc, vote)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// Params for query 'custom/gov/deposits'
type QueryDepositsParams struct {
	ProposalID int64
}

func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryDepositParams
	errRes := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	var deposits []Deposit
	depositsIterator := keeper.GetDeposits(ctx, params.ProposalID)
	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := Deposit{}
		keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), &deposit)
		deposits = append(deposits, deposit)
	}

	res, errRes = wire.MarshalJSONIndent(keeper.cdc, deposits)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// Params for query 'custom/gov/votes'
type QueryVotesParams struct {
	ProposalID int64
}

func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	var params QueryVotesParams
	errRes := keeper.cdc.UnmarshalJSON(req.Data, &params)

	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	var votes []Vote
	votesIterator := keeper.GetVotes(ctx, params.ProposalID)
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), &vote)
		votes = append(votes, vote)
	}

	res, errRes = wire.MarshalJSONIndent(keeper.cdc, votes)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}

// Params for query 'custom/gov/tally'
type QueryTallyParams struct {
	ProposalID int64
}

func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
	// TODO: Dependant on #1914

	var proposalID int64
	errRes := keeper.cdc.UnmarshalJSON(req.Data, proposalID)
	if errRes != nil {
		return []byte{}, sdk.ErrUnknownRequest(fmt.Sprintf("incorrectly formatted request data - %s", errRes.Error()))
	}

	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return []byte{}, ErrUnknownProposal(DefaultCodespace, proposalID)
	}

	var tallyResult TallyResult

	if proposal.GetStatus() == StatusDepositPeriod {
		tallyResult = EmptyTallyResult()
	} else if proposal.GetStatus() == StatusPassed || proposal.GetStatus() == StatusRejected {
		tallyResult = proposal.GetTallyResult()
	} else {
		_, tallyResult, _ = tally(ctx, keeper, proposal)
	}

	res, errRes = wire.MarshalJSONIndent(keeper.cdc, tallyResult)
	if errRes != nil {
		panic("could not marshal result to JSON")
	}
	return res, nil
}
