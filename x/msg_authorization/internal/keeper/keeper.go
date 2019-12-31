package keeper

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
)

type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	router   sdk.Router
}

// NewKeeper constructs a message authorisation Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router sdk.Router) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}
}

func (k Keeper) getActorAuthorizationKey(grantee sdk.AccAddress, granter sdk.AccAddress, msg sdk.Msg) []byte {
	return []byte(fmt.Sprintf("c/%x/%x/%s/%s", grantee, granter, msg.Route(), msg.Type()))
}

func (k Keeper) getAuthorizationGrant(ctx sdk.Context, actor []byte) (grant types.AuthorizationGrant, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(actor)
	if bz == nil {
		return grant, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &grant)
	return grant, true
}

func (k Keeper) update(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, updated types.Authorization) {
	grant, found := k.getAuthorizationGrant(ctx, k.getActorAuthorizationKey(grantee, granter, updated.MsgType()))
	if !found {
		return
	}
	grant.Authorization = updated
}

// DispatchActions attempts to execute the provided messages via authorization
// grants from the message signer to the grantee.
func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) (*sdk.Result, error) {
	var res *sdk.Result
	for _, msg := range msgs {
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "authorization can be given to msg with only one signer")
		}
		granter := signers[0]
		if !bytes.Equal(granter, grantee) {
			authorization, _ := k.GetAuthorization(ctx, grantee, granter, msg)
			if authorization == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "authorization not found")
			}
			allow, updated, del := authorization.Accept(msg, ctx.BlockHeader())
			if !allow {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "authorization not found")
			}
			if del {
				k.Revoke(ctx, grantee, granter, msg)
			} else if updated != nil {
				k.update(ctx, grantee, granter, updated)
			}
		}
		res, _ = k.router.Route(msg.Route())(ctx, msg)
		if !res.Data {
			return res
		}
	}

	return sdk.Result{}
}

// Grant method grants the provided authorization to the grantee on the granter's account with the provided expiration
// time. If there is an existing authorization grant for the same `sdk.Msg` type, this grant
// overwrites that.
func (k Keeper) Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, authorization types.Authorization, expiration time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(types.AuthorizationGrant{Authorization: authorization, Expiration: expiration})
	actor := k.getActorAuthorizationKey(grantee, granter, authorization.MsgType())
	store.Set(actor, bz)
}

// Revoke method revokes any authorization for the provided message type granted to the grantee by the granter.
func (k Keeper) Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getActorAuthorizationKey(grantee, granter, msgType))
}

// GetAuthorization Returns any `Authorization` (or `nil`), with the expiration time,
// granted to the grantee by the granter for the provided msg type.
func (k Keeper) GetAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) (cap types.Authorization, expiration time.Time) {
	grant, found := k.getAuthorizationGrant(ctx, k.getActorAuthorizationKey(grantee, granter, msgType))
	if !found {
		return nil, time.Time{}
	}
	if !grant.Expiration.IsZero() && grant.Expiration.Before(ctx.BlockHeader().Time) {
		k.Revoke(ctx, grantee, granter, msgType)
		return nil, time.Time{}
	}
	return grant.Authorization, grant.Expiration
}
