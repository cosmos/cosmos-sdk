package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type Querier struct {
	keeper Keeper
}

func NewQuerier(keeper Keeper) {
	return Querier{
		keeper: keeper,
	}
}

func (keeper Keeper) Query(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
	switch path[0] {
	case "tally":
		return QueryTally(ctx, path[1:], req)
	case "proposal":
		return handleMsgSubmitProposal(ctx, keeper, msg)
	case MsgVote:
		return handleMsgVote(ctx, keeper, msg)
	default:
		errMsg := "Unrecognized gov msg type"
		return sdk.ErrUnknownRequest(errMsg).Result()
	}
}

func QueryProposal(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
	var proposalID int64
	err := keeper.cdc.UnmarshalBinary(req.Data, proposalID)
	if err != nil {
		return []byte{}, sdk.ErrUnknownRequest()
	}
	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return []byte{}, ErrUnknownProposal(DefaultCodespace, proposalID)
	}
	return keeper.cdc.MustMarshalBinary(proposal), nil
}

func QueryTally(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
	var proposalID int64
	err := keeper.cdc.UnmarshalBinary(req.Data, proposalID)
	if err != nil {
		return []byte{}, sdk.ErrUnknownRequest()
	}
	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return []byte{}, ErrUnknownProposal(DefaultCodespace, proposalID)
	}
	passes, _ := tally(ctx, keeper, proposal)
}
