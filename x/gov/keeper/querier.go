package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewQuerier creates a new gov Querier instance
func NewQuerier(keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryProposals:
			return queryProposals(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryProposal:
			return queryProposal(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryDeposits:
			return queryDeposits(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryDeposit:
			return queryDeposit(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryVotes:
			return queryVotes(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryVote:
			return queryVote(ctx, path[1:], req, keeper, legacyQuerierCdc)

		case types.QueryTally:
			return queryTally(ctx, path[1:], req, keeper, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	switch path[0] {
	case types.ParamDeposit:
		bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, keeper.GetDepositParams(ctx))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil

	case types.ParamVoting:
		bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, keeper.GetVotingParams(ctx))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil

	case types.ParamTallying:
		bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, keeper.GetTallyParams(ctx))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
		}
		return bz, nil

	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "%s is not a valid query request path", req.Path)
	}
}

// nolint: unparam
func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	proposal, ok := keeper.GetProposal(ctx, params.ProposalID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, proposal)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// nolint: unparam
func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryDepositParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositor)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, deposit)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// nolint: unparam
func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryVoteParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, vote)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// nolint: unparam
func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	deposits := keeper.GetDeposits(ctx, params.ProposalID)
	if deposits == nil {
		deposits = types.Deposits{}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// nolint: unparam
func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	proposalID := params.ProposalID

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return nil, sdkerrors.Wrapf(types.ErrUnknownProposal, "%d", proposalID)
	}

	var tallyResult types.TallyResult

	switch {
	case proposal.Status == types.StatusDepositPeriod:
		tallyResult = types.EmptyTallyResult()

	case proposal.Status == types.StatusPassed || proposal.Status == types.StatusRejected:
		tallyResult = proposal.FinalTallyResult

	default:
		// proposal is in voting period
		_, _, tallyResult = keeper.Tally(ctx, proposal)
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, tallyResult)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// nolint: unparam
func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalVotesParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	votes := keeper.GetVotes(ctx, params.ProposalID)
	if votes == nil {
		votes = types.Votes{}
	} else {
		start, end := client.Paginate(len(votes), params.Page, params.Limit, 100)
		if start < 0 || end < 0 {
			votes = types.Votes{}
		} else {
			votes = votes[start:end]
		}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, votes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryProposals(ctx sdk.Context, _ []string, req abci.RequestQuery, keeper Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params types.QueryProposalsParams
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	proposals := keeper.GetProposalsFiltered(ctx, params)
	if proposals == nil {
		proposals = types.Proposals{}
	}

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, proposals)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
