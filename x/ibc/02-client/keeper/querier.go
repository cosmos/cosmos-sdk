package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
)

// QuerierClients defines the sdk.Querier to query all the light client states.
func QuerierClients(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllClientsParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
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
func QuerierClientState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryClientStateParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	clientState, found := k.GetClientState(ctx, params.ClientID)
	if !found {
		return nil, errors.ErrClientTypeNotFound(k.codespace)
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, clientState)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// QuerierConsensusState defines the sdk.Querier to query a consensus state
func QuerierConsensusState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryClientStateParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	consensusState, found := k.GetConsensusState(ctx, params.ClientID)
	if !found {
		return nil, errors.ErrConsensusStateNotFound(k.codespace)
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, consensusState)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// QuerierVerifiedRoot defines the sdk.Querier to query a verified commitment root
func QuerierVerifiedRoot(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryCommitmentRootParams

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	root, found := k.GetVerifiedRoot(ctx, params.ClientID, params.Height)
	if !found {
		return nil, errors.ErrRootNotFound(k.codespace)
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, root)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

// QuerierCommitter defines the sdk.Querier to query a committer
func QuerierCommitter(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryCommitterParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	committer, found := k.GetCommitter(ctx, params.ClientID, params.Height)
	if !found {
		return nil, sdk.ConvertError(errors.ErrCommitterNotFound(k.codespace,
			fmt.Sprintf("committer not found on height: %d", params.Height)))
	}

	bz, err := types.SubModuleCdc.MarshalJSON(committer)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
