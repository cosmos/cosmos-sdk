package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/subkeys/internal/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryGetFeeAllowances = "fees"
)

// NewQuerier creates a new querier
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		var res []byte
		var err error
		switch path[0] {
		case QueryGetFeeAllowances:
			res, err = queryGetFeeAllowances(ctx, path[1:], keeper)
		default:
			err = sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
				"Unknown package %s query endpoint", types.ModuleName)
		}
		return res, sdk.ConvertError(err)
	}
}

func queryGetFeeAllowances(ctx sdk.Context, args []string, keeper Keeper) ([]byte, error) {
	grantee := args[0]
	granteeAddr, err := sdk.AccAddressFromBech32(grantee)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "invalid address")
	}

	fees, err := keeper.GetAllGranteeFeeAllowances(ctx, granteeAddr)
	if err != nil {
		return nil, sdk.ConvertError(err)
	}
	if fees == nil {
		return []byte("[]"), nil
	}

	bz, err := keeper.cdc.MarshalJSON(fees)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "could not marshal query result to JSON")
	}
	return bz, nil
}
