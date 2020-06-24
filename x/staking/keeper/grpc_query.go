package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Validators(c context.Context, req *types.QueryValidatorsRequest) (*types.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	var validators types.Validators
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)

	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		result, err := types.UnmarshalValidator(k.cdc, value)
		if err != nil {
			return err
		}
		validators = append(validators, result)
		return nil
	})
	if err != nil {
		return &types.QueryValidatorsResponse{}, err
	}

	return &types.QueryValidatorsResponse{Validators: validators, Res: res}, nil

}

func (k Keeper) ValidatorQ(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	validator, found := k.GetValidator(ctx, req.ValidatorAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "validator %s not found", req.ValidatorAddr)
	}

	return &types.QueryValidatorResponse{Validator: validator}, nil
}

func (k Keeper) ValidatorDelegations(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	delegations := k.GetValidatorDelegations(ctx, req.ValidatorAddr)

	delResponses, err := DelegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Unable to convert delegations")
	}

	return &types.QueryValidatorDelegationsResponse{DelegationResponses: delResponses}, nil
}

func (k Keeper) ValidatorUnbondingDelegations(c context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorUnbondingDelegationsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)

	var ubds types.UnbondingDelegations
	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		types.GetUBDsByValIndexKey(req.ValidatorAddr)
		ubd, err := types.UnmarshalUBD(k.cdc, value)
		if err != nil {
			return err
		}
		ubds = append(ubds, ubd)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryValidatorUnbondingDelegationsResponse{UnbondingResponses: ubds, Res: res}, nil
}

func (k Keeper) DelegationQ(c context.Context, req *types.QueryBondParamsRequest) (*types.QueryDelegationResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	delegation, found := k.GetDelegation(ctx, req.DelegatorAddr, req.ValidatorAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "delegation with delegator %s not found for"+
			"valiedator %s", req.DelegatorAddr, req.ValidatorAddr)
	}

	delResponse, err := DelegationToDelegationResponse(ctx, k, delegation)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Unable to convert delegations")
	}

	return &types.QueryDelegationResponse{DelegationResponse: &delResponse}, nil
}

func (k Keeper) UnbondingDelegation(c context.Context, req *types.QueryBondParamsRequest) (*types.QueryUnbondingDelegationResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	unbond, found := k.GetUnbondingDelegation(ctx, req.DelegatorAddr, req.ValidatorAddr)
	if !found {
		return nil, status.Errorf(codes.NotFound, "unbonding delegation with delegator %s not found for"+
			"valiedator %s", req.DelegatorAddr, req.ValidatorAddr)
	}

	return &types.QueryUnbondingDelegationResponse{Unbond: unbond}, nil
}

func (k Keeper) DelegatorDelegations(c context.Context, req *types.QueryDelegatorParamsRequest) (*types.QueryDelegatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	var delegations types.Delegations
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)

	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		key = types.GetDelegationsKey(req.DelegatorAddress)
		delegation, err := types.UnmarshalDelegation(k.cdc, value)
		if err != nil {
			return err
		}
		delegations = append(delegations, delegation)
		return nil
	})

	if delegations == nil {
		return &types.QueryDelegatorDelegationsResponse{DelegationResponses: types.DelegationResponses{}}, nil
	}
	delegationResps, err := DelegationsToDelegationResponses(ctx, k, delegations)
	if err != nil {
		return nil, err
	}

	return &types.QueryDelegatorDelegationsResponse{DelegationResponses: delegationResps, Res: res}, nil

}

func (k Keeper) DelegatorValidator(c context.Context, req *types.QueryBondParamsRequest) (*types.QueryDelegatorValidatorResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	validator, err := k.GetDelegatorValidator(ctx, req.DelegatorAddr, req.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryDelegatorValidatorResponse{Validator: validator}, nil
}

func (k Keeper) DelegatorUnbondingDelegations(c context.Context, req *types.QueryDelegatorParamsRequest) (*types.QueryUnbondingDelegatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	var unbondingDelegations types.UnbondingDelegations
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		key = types.GetUBDsKey(req.DelegatorAddress)
		unbond, err := types.UnmarshalUBD(k.cdc, value)
		if err != nil {
			return err
		}
		unbondingDelegations = append(unbondingDelegations, unbond)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryUnbondingDelegatorDelegationsResponse{UnbondingResponses: unbondingDelegations, Res: res}, nil
}

func (k Keeper) HistoricalInfo(c context.Context, req *types.QueryHistoricalInfoRequest) (*types.QueryHistoricalInfoResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	hi, found := k.GetHistoricalInfo(ctx, req.Height)
	if !found {
		return nil, types.ErrNoHistoricalInfo
	}

	return &types.QueryHistoricalInfoResponse{Hist: &hi}, nil
}

func (k Keeper) Pool(c context.Context, _ *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	bondDenom := k.BondDenom(ctx)
	bondedPool := k.GetBondedPool(ctx)
	notBondedPool := k.GetNotBondedPool(ctx)

	if bondedPool == nil || notBondedPool == nil {
		return nil, status.Errorf(codes.FailedPrecondition, "pool accounts haven't been set")
	}

	pool := types.NewPool(
		k.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom).Amount,
		k.bankKeeper.GetBalance(ctx, bondedPool.GetAddress(), bondDenom).Amount,
	)

	return &types.QueryPoolResponse{Pool: pool}, nil
}

func (k Keeper) Parameters(c context.Context, _ *types.QueryParametersRequest) (*types.QueryParametersResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)

	return &types.QueryParametersResponse{Params: params}, nil
}
