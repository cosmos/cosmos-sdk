package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

const (
	QueryGetFeeAllowances = "fees"
)

// NewQuerier creates a new querier
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		var (
			res []byte
			err error
		)

		switch path[0] {
		case QueryGetFeeAllowances:
			res, err = queryGetFeeAllowances(ctx, path[1:], keeper)

		default:
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}

		return res, err
	}
}

func queryGetFeeAllowances(ctx sdk.Context, args []string, keeper Keeper) ([]byte, error) {
	grantee := args[0]
	granteeAddr, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid address")
	}

	var grants []types.FeeAllowanceGrant
	err = keeper.IterateAllGranteeFeeAllowances(ctx, granteeAddr, func(grant types.FeeAllowanceGrant) bool {
		grants = append(grants, grant)
		return false
	})
	if err != nil {
		return nil, err
	}

	if grants == nil {
		return []byte("[]"), nil
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, grants)
	//bz, err := keeper.cdc.MarshalJSON(grants)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
