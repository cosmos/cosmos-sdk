package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/codec"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	sdkerrors "github.com/KiraCore/cosmos-sdk/types/errors"
	"github.com/KiraCore/cosmos-sdk/x/params/types"
	"github.com/KiraCore/cosmos-sdk/x/params/types/proposal"
)

// NewQuerier returns a new querier handler for the x/params module.
func NewQuerier(k Keeper, legacyQuerierCdc codec.JSONMarshaler) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, req, k, legacyQuerierCdc)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, req abci.RequestQuery, k Keeper, legacyQuerierCdc codec.JSONMarshaler) ([]byte, error) {
	var params types.QuerySubspaceParams

	if err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	ss, ok := k.GetSubspace(params.Subspace)
	if !ok {
		return nil, sdkerrors.Wrap(proposal.ErrUnknownSubspace, params.Subspace)
	}

	rawValue := ss.GetRaw(ctx, []byte(params.Key))
	resp := types.NewSubspaceParamsResponse(params.Subspace, params.Key, string(rawValue))

	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, resp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
