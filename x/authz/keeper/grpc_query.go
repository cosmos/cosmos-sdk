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
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ authz.QueryServer = Keeper{}

// Authorizations implements the Query/Grants gRPC method.
func (k Keeper) Grants(c context.Context, req *authz.QueryGrantsRequest) (*authz.QueryGrantsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	granter, err := sdk.AccAddressFromBech32(req.Granter)
	if err != nil {
		return nil, err
	}

	grantee, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	key := grantStoreKey(grantee, granter, "")
	authStore := prefix.NewStore(store, key)

	if req.MsgTypeUrl != "" {
		authorization, expiration := k.GetCleanAuthorization(ctx, grantee, granter, req.MsgTypeUrl)
		if authorization == nil {
			return nil, status.Errorf(codes.NotFound, "no authorization found for %s type", req.MsgTypeUrl)
		}
		authorizationAny, err := codectypes.NewAnyWithValue(authorization)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		return &authz.QueryGrantsResponse{
			Grants: []*authz.Grant{{
				Authorization: authorizationAny,
				Expiration:    expiration,
			}},
		}, nil
	}

	var authorizations []*authz.Grant
	pageRes, err := query.FilteredPaginate(authStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		auth, err := unmarshalAuthorization(k.cdc, value)
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
			authorizations = append(authorizations, &authz.Grant{
				Authorization: authorizationAny,
				Expiration:    auth.Expiration,
			})
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &authz.QueryGrantsResponse{
		Grants:     authorizations,
		Pagination: pageRes,
	}, nil
}

// GranterGrants implements the Query/GranterGrants gRPC method.
func (k Keeper) GranterGrants(c context.Context, req *authz.QueryGranterGrantsRequest) (*authz.QueryGranterGrantsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	granter, err := sdk.AccAddressFromBech32(req.Granter)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := ctx.KVStore(k.storeKey)
	authzStore := prefix.NewStore(store, grantStoreKey(nil, granter, ""))

	var grants []*authz.GrantAuthorization
	pageRes, err := query.FilteredPaginate(authzStore, req.Pagination, func(key []byte, value []byte,
		accumulate bool) (bool, error) {
		auth, err := unmarshalAuthorization(k.cdc, value)
		if err != nil {
			return false, err
		}

		auth1 := auth.GetAuthorization()
		if accumulate {
			any, err := codectypes.NewAnyWithValue(auth1)
			if err != nil {
				return false, status.Errorf(codes.Internal, err.Error())
			}

			grantee := firstAddressFromGrantStoreKey(key)
			grants = append(grants, &authz.GrantAuthorization{
				Granter:       granter.String(),
				Grantee:       grantee.String(),
				Authorization: any,
				Expiration:    auth.Expiration,
			})
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &authz.QueryGranterGrantsResponse{
		Grants:     grants,
		Pagination: pageRes,
	}, nil
}

// GranteeGrants implements the Query/GranteeGrants gRPC method.
func (k Keeper) GranteeGrants(c context.Context, req *authz.QueryGranteeGrantsRequest) (*authz.QueryGranteeGrantsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	grantee, err := sdk.AccAddressFromBech32(req.Grantee)
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), GrantKey)

	var authorizations []*authz.GrantAuthorization
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(key []byte, value []byte,
		accumulate bool) (bool, error) {
		auth, err := unmarshalAuthorization(k.cdc, value)
		if err != nil {
			return false, err
		}

		granter, g := addressesFromGrantStoreKey(append(GrantKey, key...))
		if !g.Equals(grantee) {
			return false, nil
		}

		auth1 := auth.GetAuthorization()
		if accumulate {
			any, err := codectypes.NewAnyWithValue(auth1)
			if err != nil {
				return false, status.Errorf(codes.Internal, err.Error())
			}

			authorizations = append(authorizations, &authz.GrantAuthorization{
				Authorization: any,
				Expiration:    auth.Expiration,
				Granter:       granter.String(),
				Grantee:       grantee.String(),
			})
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &authz.QueryGranteeGrantsResponse{
		Grants:     authorizations,
		Pagination: pageRes,
	}, nil
}

// unmarshal an authorization from a store value
func unmarshalAuthorization(cdc codec.BinaryCodec, value []byte) (v authz.Grant, err error) {
	err = cdc.Unmarshal(value, &v)
	return v, err
}
