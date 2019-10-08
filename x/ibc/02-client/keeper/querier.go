package keeper

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// TODO: return proof

// NewQuerier creates a querier for the IBC client
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
		switch path[0] {
		case types.QueryClientState:
			return queryClientState(ctx, req, k) // TODO: return proof
		case types.QueryConsensusState:
			return queryConsensusState(ctx, req, k) // TODO: return proof
		case types.QueryCommitmentPath:
			return queryCommitmentPath(k)
		case types.QueryCommitmentRoot:
			return queryCommitmentRoot(ctx, req, k) // TODO: return proof
		// case types.QueryHeader:
		// 	return queryHeader(ctx, req, k)
		default:
			return nil, sdk.ErrUnknownRequest("unknown IBC client query endpoint")
		}
	}
}

func queryClientState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryClientStateParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	// NOTE: clientID won't be exported as it's declared as private
	// TODO: should we create a custom ExportedClientState to make it public ?
	clientState, found := k.GetClientState(ctx, params.ClientID)
	if !found {
		return nil, types.ErrClientTypeNotFound(k.codespace)
	}

	bz, err := types.SubModuleCdc.MarshalJSON(clientState)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryConsensusState(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryClientStateParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	consensusState, found := k.GetConsensusState(ctx, params.ClientID)
	if !found {
		return nil, types.ErrConsensusStateNotFound(k.codespace)
	}

	bz, err := types.SubModuleCdc.MarshalJSON(consensusState)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryCommitmentPath(k Keeper) ([]byte, sdk.Error) {
	path := k.GetCommitmentPath()

	bz, err := types.SubModuleCdc.MarshalJSON(path)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

func queryCommitmentRoot(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {
	var params types.QueryCommitmentRootParams

	err := types.SubModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	root, found := k.GetCommitmentRoot(ctx, params.ClientID, params.Height)
	if !found {
		return nil, types.ErrRootNotFound(k.codespace)
	}

	bz, err := types.SubModuleCdc.MarshalJSON(root)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// TODO: this is implented directly on the client
// func queryHeader(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, sdk.Error) {

// 	res, err := codec.MarshalJSONIndent(types.SubModuleCdc, header)
// 	if err != nil {
// 		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("failed to marshal result to JSON", err.Error()))
// 	}

// 	return res, nil
// }
