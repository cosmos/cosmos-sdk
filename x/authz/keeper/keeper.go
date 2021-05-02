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
	cdc      codec.BinaryCodec
	router   *baseapp.MsgServiceRouter
}

// NewKeeper constructs a message authorization Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc codec.BinaryCodec, router *baseapp.MsgServiceRouter) Keeper {
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
func (k Keeper) getAuthorizationGrant(ctx sdk.Context, grantStoreKey []byte) (grant types.AuthorizationGrant, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(grantStoreKey)
	if bz == nil {
		return grant, false
	}
	k.cdc.MustUnmarshal(bz, &grant)
	return grant, true
}

func (k Keeper) update(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, updated exported.Authorization) error {
	grantStoreKey := types.GetAuthorizationStoreKey(grantee, granter, updated.MethodName())
	grant, found := k.getAuthorizationGrant(ctx, grantStoreKey)
	if !found {
		return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "authorization not found")
	}

	msg, ok := updated.(proto.Message)
	if !ok {
		sdkerrors.Wrapf(sdkerrors.ErrPackAny, "cannot proto marshal %T", updated)
	}

	any, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return err
	}

	grant.Authorization = any
	store := ctx.KVStore(k.storeKey)
	store.Set(grantStoreKey, k.cdc.MustMarshal(&grant))
	return nil
}

// DispatchActions attempts to execute the provided messages via authorization
// grants from the message signer to the grantee.
func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) (*sdk.Result, error) {
	var msgResult *sdk.Result
	var err error
	for _, msg := range msgs {
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "authorization can be given to msg with only one signer")
		}
		granter := signers[0]
		if !granter.Equals(grantee) {
			authorization, _ := k.GetOrRevokeAuthorization(ctx, grantee, granter, sdk.MsgTypeURL(msg))
			if authorization == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "authorization not found")
			}
			updated, del, err := authorization.Accept(ctx, msg)
			if err != nil {
				return nil, err
			}
			if del {
				err = k.Revoke(ctx, grantee, granter, sdk.MsgTypeURL(msg))
				if err != nil {
					return nil, err
				}
			} else if updated != nil {
				err = k.update(ctx, grantee, granter, updated)
				if err != nil {
					return nil, err
				}
			}
		}
		handler := k.router.Handler(msg)

		if handler == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s", sdk.MsgTypeURL(msg))
		}

		msgResult, err = handler(ctx, msg)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message %v", msg)
		}
	}

	return msgResult, nil
}

// Grant method grants the provided authorization to the grantee on the granter's account with the provided expiration
// time. If there is an existing authorization grant for the same `sdk.Msg` type, this grant
// overwrites that.
func (k Keeper) Grant(ctx sdk.Context, grantee, granter sdk.AccAddress, authorization exported.Authorization, expiration time.Time) error {
	store := ctx.KVStore(k.storeKey)

	grant, err := types.NewAuthorizationGrant(authorization, expiration)
	if err != nil {
		return err
	}

	bz := k.cdc.MustMarshal(&grant)
	grantStoreKey := types.GetAuthorizationStoreKey(grantee, granter, authorization.MethodName())
	store.Set(grantStoreKey, bz)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventGrantAuthorization,
			sdk.NewAttribute(types.AttributeKeyGrantType, authorization.MethodName()),
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, granter.String()),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, grantee.String()),
		),
	)
	return nil
}

// Revoke method revokes any authorization for the provided message type granted to the grantee by the granter.
func (k Keeper) Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) error {
	store := ctx.KVStore(k.storeKey)
	grantStoreKey := types.GetAuthorizationStoreKey(grantee, granter, msgType)
	_, found := k.getAuthorizationGrant(ctx, grantStoreKey)
	if !found {
		return sdkerrors.Wrap(sdkerrors.ErrNotFound, "authorization not found")
	}
	store.Delete(grantStoreKey)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventRevokeAuthorization,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyGrantType, msgType),
			sdk.NewAttribute(types.AttributeKeyGranterAddress, granter.String()),
			sdk.NewAttribute(types.AttributeKeyGranteeAddress, grantee.String()),
		),
	)
	return nil
}

// GetAuthorizations Returns list of `Authorizations` granted to the grantee by the granter.
func (k Keeper) GetAuthorizations(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress) (authorizations []exported.Authorization) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetAuthorizationStoreKey(grantee, granter, "")
	iter := sdk.KVStorePrefixIterator(store, key)
	defer iter.Close()
	var authorization types.AuthorizationGrant
	for ; iter.Valid(); iter.Next() {
		k.cdc.MustUnmarshal(iter.Value(), &authorization)
		authorizations = append(authorizations, authorization.GetAuthorizationGrant())
	}
	return authorizations
}

// GetOrRevokeAuthorization Returns any `Authorization` (or `nil`), with the expiration time,
// granted to the grantee by the granter for the provided msg type.
// If the Authorization is expired already, it will revoke the authorization and return nil
func (k Keeper) GetOrRevokeAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (cap exported.Authorization, expiration time.Time) {
	grant, found := k.getAuthorizationGrant(ctx, types.GetAuthorizationStoreKey(grantee, granter, msgType))
	if !found {
		return nil, time.Time{}
	}
	if grant.Expiration.Before(ctx.BlockHeader().Time) {
		k.Revoke(ctx, grantee, granter, msgType)
		return nil, time.Time{}
	}

	return grant.GetAuthorizationGrant(), grant.Expiration
}

// IterateGrants iterates over all authorization grants
func (k Keeper) IterateGrants(ctx sdk.Context,
	handler func(granterAddr sdk.AccAddress, granteeAddr sdk.AccAddress, grant types.AuthorizationGrant) bool) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.GrantKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var grant types.AuthorizationGrant
		granterAddr, granteeAddr := types.ExtractAddressesFromGrantKey(iter.Key())
		k.cdc.MustUnmarshal(iter.Value(), &grant)
		if handler(granterAddr, granteeAddr, grant) {
			break
		}
	}
}
