package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
)

// QueryAllClients defines the sdk.Querier to query all the light client states.
func QueryAllClients(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllClientsParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	clients := k.GetAllClients(ctx)

	start, end := client.Paginate(len(clients), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		clients = []types.State{}
	} else {
		clients = clients[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, clients)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

// QuerierClientState defines the sdk.Querier to query the IBC client state
func QuerierClientState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryClientStateParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	clientState, found := k.GetClientState(ctx, params.ClientID)
	if !found {
		return nil, sdk.ConvertError(errors.ErrClientTypeNotFound(k.codespace))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(clientState)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// QuerierConsensusState defines the sdk.Querier to query a consensus state
func QuerierConsensusState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryClientStateParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	consensusState, found := k.GetConsensusState(ctx, params.ClientID)
	if !found {
		return nil, sdk.ConvertError(errors.ErrConsensusStateNotFound(k.codespace))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(consensusState)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// QuerierVerifiedRoot defines the sdk.Querier to query a verified commitment root
func QuerierVerifiedRoot(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryCommitmentRootParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	root, found := k.GetVerifiedRoot(ctx, params.ClientID, params.Height)
	if !found {
		return nil, sdk.ConvertError(errors.ErrRootNotFound(k.codespace))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(root)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
