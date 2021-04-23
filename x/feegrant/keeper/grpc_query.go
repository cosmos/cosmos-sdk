package keeper

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/feegrant/types"
)

var _ types.QueryServer = Keeper{}

// FeeAllowance returns fee granted to the grantee by the granter.
func (q Keeper) FeeAllowance(c context.Context, req *types.QueryFeeAllowanceRequest) (*types.QueryFeeAllowanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	granterAddr, err := sdk.AccAddressFromBech32(req.Granter)
	if err != nil {
		return nil, err
	}

	granteeAddr, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	feeAllowance, err := q.GetFeeAllowance(ctx, granterAddr, granteeAddr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	msg, ok := feeAllowance.(proto.Message)
	if !ok {
		return nil, status.Errorf(codes.Internal, "can't proto marshal %T", msg)
	}

	feeAllowanceAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryFeeAllowanceResponse{
		FeeAllowance: &types.FeeAllowanceGrant{
			Granter:   granterAddr.String(),
			Grantee:   granteeAddr.String(),
			Allowance: feeAllowanceAny,
		},
	}, nil
}

// FeeAllowances queries all the allowances granted to the given grantee.
func (q Keeper) FeeAllowances(c context.Context, req *types.QueryFeeAllowancesRequest) (*types.QueryFeeAllowancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	granteeAddr, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	var grants []*types.FeeAllowanceGrant

	store := ctx.KVStore(q.storeKey)
	grantsStore := prefix.NewStore(store, types.FeeAllowancePrefixByGrantee(granteeAddr))

	pageRes, err := query.Paginate(grantsStore, req.Pagination, func(key []byte, value []byte) error {
		var grant types.FeeAllowanceGrant

		if err := q.cdc.UnmarshalBinaryBare(value, &grant); err != nil {
			return err
		}

		grants = append(grants, &grant)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryFeeAllowancesResponse{FeeAllowances: grants, Pagination: pageRes}, nil
}
