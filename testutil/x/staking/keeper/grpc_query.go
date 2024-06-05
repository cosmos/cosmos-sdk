package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/x/staking/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

func NewQuerier(keeper *Keeper) Querier {
	return Querier{Keeper: keeper}
}

// Validators queries all validators that match the given status
func (k Querier) Validators(ctx context.Context, req *types.QueryValidatorsRequest) (*types.QueryValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// validate the provided status, return all the validators if the status is empty
	if req.Status != "" && !(req.Status == types.Bonded.String() || req.Status == types.Unbonded.String() || req.Status == types.Unbonding.String()) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid validator status %s", req.Status)
	}

	store := runtime.KVStoreAdapter(k.KVStoreService.OpenKVStore(ctx))
	valStore := prefix.NewStore(store, types.ValidatorsKey)

	validators, pageRes, err := query.GenericFilteredPaginate(k.cdc, valStore, req.Pagination, func(key []byte, val *types.Validator) (*types.Validator, error) {
		if req.Status != "" && !strings.EqualFold(val.GetStatus().String(), req.Status) {
			return nil, nil
		}

		return val, nil
	}, func() *types.Validator {
		return &types.Validator{}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	vals := types.Validators{}
	for _, val := range validators {
		vals.Validators = append(vals.Validators, *val)
	}

	return &types.QueryValidatorsResponse{Validators: vals.Validators, Pagination: pageRes}, nil
}

// Validator queries validator info for given validator address
func (k Querier) Validator(ctx context.Context, req *types.QueryValidatorRequest) (*types.QueryValidatorResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.ValidatorAddr == "" {
		return nil, status.Error(codes.InvalidArgument, "validator address cannot be empty")
	}

	valAddr, err := k.validatorAddressCodec.StringToBytes(req.ValidatorAddr)
	if err != nil {
		return nil, err
	}

	validator, err := k.GetValidator(ctx, valAddr)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "validator %s not found", req.ValidatorAddr)
	}

	return &types.QueryValidatorResponse{Validator: validator}, nil
}
