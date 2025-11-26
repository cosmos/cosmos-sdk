package keeper

import (
	"context"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/x/consensus/exported"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

var _ exported.Authority = Keeper{}

func (k Keeper) GetAuthority(ctx context.Context) (string, error) {
	params, err := k.Params(ctx, &consensusparamtypes.QueryParamsRequest{})
	if err != nil {
		return "", err
	}
	return params.Params.Authority.Authority, err
}

func (k Keeper) ValidateAuthority(ctx context.Context, address string) error {
	params, err := k.Params(ctx, &consensusparamtypes.QueryParamsRequest{})
	if err != nil {
		return err
	}
	if params.Params.Authority.Authority != address {
		return errors.Wrapf(consensusparamtypes.ErrUnauthorized, "expected %s for authority, got %s", params.Params.Authority.Authority, address)
	}
	return nil
}
