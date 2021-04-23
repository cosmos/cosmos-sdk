package keeper

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      codec.BinaryMarshaler
	router   *baseapp.MsgServiceRouter
}

// NewKeeper constructs a message authorization Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc codec.BinaryMarshaler, router *baseapp.MsgServiceRouter) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// getAuthorizationGrant returns grant between granter and grantee for the given msg type
func (k Keeper) getAuthorizationGrant(ctx sdk.Context, grantStoreKey []byte) (grant types.Grant, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(grantStoreKey)
	if bz == nil {
		return grant, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &grant)
	return grant, true
}

func (k Keeper) update(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, updated exported.Authorization) error {
	grantStoreKey := types.GetAuthorizationStoreKey(grantee, granter, updated.MethodName())
	grant, found := k.getAuthorizationGrant(ctx, grantStoreKey)
	if !found {
		return sdkerrors.ErrNotFound.Wrap("authorization not found")
	}

	msg, ok := updated.(proto.Message)
	if !ok {
		sdkerrors.ErrPackAny.Wrapf("cannot proto marshal %T", updated)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}

	grant.Authorization = any
	store := ctx.KVStore(k.storeKey)
	store.Set(grantStoreKey, k.cdc.MustMarshalBinaryBare(&grant))
	return nil
}

// DispatchActions attempts to execute the provided messages via authorization
// grants from the message signer to the grantee.
func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, serviceMsgs []sdk.ServiceMsg) (*sdk.Result, error) {
	var msgResult *sdk.Result
	var err error
	for _, serviceMsg := range serviceMsgs {
		signers := serviceMsg.GetSigners()
		if len(signers) != 1 {
			return nil, sdkerrors.ErrInvalidRequest.Wrap("authorization can be given to msg with only one signer")
		}
		granter := signers[0]
		// if granter != grantee then check authorization.Accept, otherwise we implicitly accept.
		if !granter.Equals(grantee) {
			authorization, _ := k.GetOrRevokeAuthorization(ctx, grantee, granter, serviceMsg.MethodName)
			if authorization == nil {
				return nil, sdkerrors.ErrUnauthorized.Wrap("authorization not found")
			}
			resp, err := authorization.Accept(ctx, serviceMsg)
			if err != nil {
				return nil, err
			}
			if resp.Delete {
				k.DeleteGrant(ctx, grantee, granter, serviceMsg.Type())
			} else if resp.Updated != nil {
				err = k.update(ctx, grantee, granter, resp.Updated)
				if err != nil {
					return nil, err
				}
			}
			if !resp.Accept {
				return nil, sdkerrors.ErrUnauthorized
			}
		}
		handler := k.router.Handler(serviceMsg.Route())

		if handler == nil {
			return nil, sdkerrors.ErrUnknownRequest.Wrapf("unrecognized message route: %s", serviceMsg.Route())
		}

		msgResult, err = handler(ctx, serviceMsg.Request)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message %s", serviceMsg.MethodName)
		}
	}

	return msgResult, nil
}

// SaveGrant method grants the provided authorization to the grantee on the granter's account
// with the provided expiration time. If there is an existing authorization grant for the
// same `sdk.Msg` type, this grant overwrites that.
func (k Keeper) SaveGrant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization exported.Authorization, expiration time.Time) error {
	store := ctx.KVStore(k.storeKey)

	grant, err := types.NewGrant(authorization, expiration)
	if err != nil {
		return err
	}

	bz := k.cdc.MustMarshalBinaryBare(&grant)
	grantStoreKey := types.GetAuthorizationStoreKey(grantee, granter, authorization.MethodName())
	store.Set(grantStoreKey, bz)
	return ctx.EventManager().EmitTypedEvent(&types.EventGrant{
		Module:  types.ModuleName,
		Msg:     authorization.MethodName(),
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
}

// DeleteGrant revokes any authorization for the provided message type granted to the grantee
// by the granter.
func (k Keeper) DeleteGrant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error {
	store := ctx.KVStore(k.storeKey)
	grantStoreKey := types.GetAuthorizationStoreKey(grantee, granter, msgType)
	_, found := k.getAuthorizationGrant(ctx, grantStoreKey)
	if !found {
		return sdkerrors.ErrNotFound.Wrap("authorization not found")
	}
	store.Delete(grantStoreKey)
	return ctx.EventManager().EmitTypedEvent(&types.EventRevoke{
		Module:  types.ModuleName,
		Msg:     msgType,
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
}

// GetAuthorizations Returns list of `Authorizations` granted to the grantee by the granter.
func (k Keeper) GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) (authorizations []exported.Authorization) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuthorizationStoreKey(grantee, granter, "")
	iter := sdk.KVStorePrefixIterator(store, key)
	defer iter.Close()
	var authorization types.Grant
	for ; iter.Valid(); iter.Next() {
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &authorization)
		authorizations = append(authorizations, authorization.GetAuthorization())
	}
	return authorizations
}

// GetOrRevokeAuthorization returns an `Authorization` and it's expiration time if the granter
// has a grant for (granter, message name) pair. If there is no grant `nil` is returned.
// If the grant is expired, the grant is revoked, removed from the storage, and `nil` is returned.
func (k Keeper) GetOrRevokeAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (cap exported.Authorization, expiration time.Time) {
	grant, found := k.getAuthorizationGrant(ctx, types.GetAuthorizationStoreKey(grantee, granter, msgType))
	if !found {
		return nil, time.Time{}
	}
	if grant.Expiration.Before(ctx.BlockHeader().Time) {
		k.DeleteGrant(ctx, grantee, granter, msgType)
		return nil, time.Time{}
	}

	return grant.GetAuthorization(), grant.Expiration
}

// IterateGrants iterates over all authorization grants
func (k Keeper) IterateGrants(ctx sdk.Context,
	handler func(granterAddr sdk.AccAddress, granteeAddr sdk.AccAddress, grant types.Grant) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GrantKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var grant types.Grant
		granterAddr, granteeAddr := types.ExtractAddressesFromGrantKey(iter.Key())
		k.cdc.MustUnmarshalBinaryBare(iter.Value(), &grant)
		if handler(granterAddr, granteeAddr, grant) {
			break
		}
	}
}
