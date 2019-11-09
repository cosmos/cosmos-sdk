package keeper

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case types.QueryEvidence:
			res, err = queryEvidence(ctx, path[1:], req, k)

		case types.QueryAllEvidence:
			res, err = queryAllEvidence(ctx, path[1:], req, k)

		default:
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint", types.ModuleName)
		}

		return res, sdk.ConvertError(err)
	}
}

func queryEvidence(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryEvidenceParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	hash, err := hex.DecodeString(params.EvidenceHash)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to decode evidence hash string query")
	}

	evidence, ok := k.GetEvidence(ctx, hash)
	if !ok {
		return nil, types.ErrNoEvidenceExists(k.codespace, params.EvidenceHash)
	}

	res, err := codec.MarshalJSONIndent(k.cdc, evidence)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryAllEvidence(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryAllEvidenceParams

	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	evidence := k.GetAllEvidence(ctx)

	start, end := client.Paginate(len(evidence), params.Page, params.Limit, 100)
	if start < 0 || end < 0 {
		evidence = []exported.Evidence{}
	} else {
		evidence = evidence[start:end]
	}

	res, err := codec.MarshalJSONIndent(k.cdc, evidence)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}
