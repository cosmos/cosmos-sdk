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
	router   types.Router
}

// NewKeeper constructs a message authorisation Keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router types.Router) Keeper {
	return Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}
}

func (k Keeper) getActorAuthorizationKey(grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) []byte {
	return []byte(fmt.Sprintf("c/%x/%x/%s", grantee, granter, msgType))
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
	var msgResult *sdk.Result
	var err error
	for _, msg := range msgs {
		fmt.Printf("this is %v", msg)
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "authorization can be given to msg with only one signer")
		}
		granter := signers[0]
		if !bytes.Equal(granter, grantee) {
			authorization, _ := k.GetAuthorization(ctx, grantee, granter, msg.Type())
			if authorization == nil {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "authorization not found")
			}
			allow, updated, del := authorization.Accept(msg, ctx.BlockHeader())
			fmt.Println("inside accept: ", allow, updated, del)
			if !allow {
				return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "authorization not found")
			}
			if del {
				k.Revoke(ctx, grantee, granter, msg.Type())
			} else if updated != nil {
				k.update(ctx, grantee, granter, updated)
			}
		}
		handler := k.router.Route(ctx, msg.Route())
		if handler == nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized message route: %s", msg.Route())
		}

		msgResult, err = handler(ctx, msg)
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "failed to execute message; message %s", msg.Type())
		}
	}

	return msgResult, nil
}

// Grant method grants the provided authorization to the grantee on the granter's account with the provided expiration
// time. If there is an existing authorization grant for the same `sdk.Msg` type, this grant
// overwrites that.
func (k Keeper) Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, authorization types.Authorization, expiration time.Time) bool {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(types.AuthorizationGrant{Authorization: authorization, Expiration: expiration.Unix()})
	actor := k.getActorAuthorizationKey(grantee, granter, authorization.MsgType())
	store.Set(actor, bz)
	return true
}

// Revoke method revokes any authorization for the provided message type granted to the grantee by the granter.
func (k Keeper) Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) bool {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getActorAuthorizationKey(grantee, granter, msgType))
	return true
}

// GetAuthorization Returns any `Authorization` (or `nil`), with the expiration time,
// granted to the grantee by the granter for the provided msg type.
func (k Keeper) GetAuthorization(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType string) (cap types.Authorization, expiration int64) {
	grant, found := k.getAuthorizationGrant(ctx, k.getActorAuthorizationKey(grantee, granter, msgType))
	if !found {
		return nil, 0
	}
	if grant.Expiration != 0 && grant.Expiration < (ctx.BlockHeader().Time.Unix()) {
		k.Revoke(ctx, grantee, granter, msgType)
		return nil, 0
	}
	return grant.Authorization, grant.Expiration
}
