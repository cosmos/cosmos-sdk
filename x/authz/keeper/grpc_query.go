package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/gogo/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

var _ types.QueryServer = Keeper{}

// Authorizations implements the Query/Authorizations gRPC method.
func (k Keeper) Authorizations(c context.Context, req *types.QueryAuthorizationsRequest) (*types.QueryAuthorizationsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.GranterAddr == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid granter addr")
	}

	if req.GranteeAddr == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid grantee addr")
	}

	granterAddr, err := sdk.AccAddressFromBech32(req.GranterAddr)

	if err != nil {
		return nil, err
	}
	granteeAddr, err := sdk.AccAddressFromBech32(req.GranteeAddr)

	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	key := types.GetActorAuthorizationKey(granteeAddr, granterAddr, "")
	authStore := prefix.NewStore(store, key)
	var authorizations []*types.AuthorizationGrant
	pageRes, err := query.FilteredPaginate(authStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		auth, err := UnmarshalAuthorization(k.cdc, value)
		if err != nil {
			return false, err
		}
		auth1 := auth.GetAuthorization()
		if accumulate {
			msg, ok := auth1.(proto.Message)
			if !ok {
				return false, status.Errorf(codes.Internal, "can't protomarshal %T", msg)
			}

			authorizationAny, err := codectypes.NewAnyWithValue(msg)
			if err != nil {
				return false, status.Errorf(codes.Internal, err.Error())
			}
			authorizations = append(authorizations, &types.AuthorizationGrant{
				Authorization: authorizationAny,
				Expiration:    auth.Expiration,
			})
		}
		return true, nil
	})

	return &types.QueryAuthorizationsResponse{
		Authorizations: authorizations,
		Pagination:     pageRes,
	}, nil
}

// unmarshal an authorization from a store value
func UnmarshalAuthorization(cdc codec.BinaryMarshaler, value []byte) (v types.AuthorizationGrant, err error) {
	err = cdc.UnmarshalBinaryBare(value, &v)
	return v, err
}

// Authorization implements the Query/Authorization gRPC method.
func (k Keeper) Authorization(c context.Context, req *types.QueryAuthorizationRequest) (*types.QueryAuthorizationResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if req.GranterAddr == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid granter addr")
	}

	if req.GranteeAddr == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid grantee addr")
	}

	if req.MsgType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "invalid msg-type")
	}

	granterAddr, err := sdk.AccAddressFromBech32(req.GranterAddr)

	if err != nil {
		return nil, err
	}
	granteeAddr, err := sdk.AccAddressFromBech32(req.GranteeAddr)

	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(c)

	authorization, expiration := k.GetAuthorization(ctx, granteeAddr, granterAddr, req.MsgType)
	if authorization == nil {
		return nil, status.Errorf(codes.NotFound, "no authorization found for %s type", req.MsgType)
	}

	msg, ok := authorization.(proto.Message)
	if !ok {
		return nil, status.Errorf(codes.Internal, "can't protomarshal %T", msg)
	}

	authorizationAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &types.QueryAuthorizationResponse{
		Authorization: &types.AuthorizationGrant{
			Authorization: authorizationAny,
			Expiration:    expiration,
		},
	}, nil
}
