package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	store := ctx.KVStore(k.storeKey)
	key := grantStoreKey(grantee, granter, "")
	grantsStore := prefix.NewStore(store, key)

	authorizations, pageRes, err := query.GenericFilteredPaginate(k.cdc, grantsStore, req.Pagination, func(key []byte, auth *authz.Grant) (*authz.Grant, error) {
		auth1 := auth.GetAuthorization()
		if err != nil {
			return nil, err
		}

		authorizationAny, err := codectypes.NewAnyWithValue(auth1)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		return &authz.Grant{
			Authorization: authorizationAny,
			Expiration:    auth.Expiration,
		}, nil
	}, func() *authz.Grant {
		return &authz.Grant{}
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

	grants, pageRes, err := query.GenericFilteredPaginate(k.cdc, authzStore, req.Pagination, func(key []byte, auth *authz.Grant) (*authz.GrantAuthorization, error) {
		auth1 := auth.GetAuthorization()
		if err != nil {
			return nil, err
		}

		any, err := codectypes.NewAnyWithValue(auth1)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		grantee := firstAddressFromGrantStoreKey(key)
		return &authz.GrantAuthorization{
			Granter:       granter.String(),
			Grantee:       grantee.String(),
			Authorization: any,
			Expiration:    auth.Expiration,
		}, nil

	}, func() *authz.Grant {
		return &authz.Grant{}
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

	authorizations, pageRes, err := query.GenericFilteredPaginate(k.cdc, store, req.Pagination, func(key []byte, auth *authz.Grant) (*authz.GrantAuthorization, error) {
		auth1 := auth.GetAuthorization()
		if err != nil {
			return nil, err
		}

		granter, g := addressesFromGrantStoreKey(append(GrantKey, key...))
		if !g.Equals(grantee) {
			return nil, nil
		}

		authorizationAny, err := codectypes.NewAnyWithValue(auth1)
		if err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}

		return &authz.GrantAuthorization{
			Authorization: authorizationAny,
			Expiration:    auth.Expiration,
			Granter:       granter.String(),
			Grantee:       grantee.String(),
		}, nil
	}, func() *authz.Grant {
		return &authz.Grant{}
	})
	if err != nil {
		return nil, err
	}

	return &authz.QueryGranteeGrantsResponse{
		Grants:     authorizations,
		Pagination: pageRes,
	}, nil
}
