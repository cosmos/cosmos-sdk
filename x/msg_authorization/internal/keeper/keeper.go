package keeper

import (
	"bytes"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (k Keeper) getActorCapabilityKey(grantee sdk.AccAddress, granter sdk.AccAddress, msg sdk.Msg) []byte {
	return []byte(fmt.Sprintf("c/%x/%x/%s/%s", grantee, granter, msg.Route(), msg.Type()))
}

func (k Keeper) getCapabilityGrant(ctx sdk.Context, actor []byte) (grant types.CapabilityGrant, found bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(actor)
	if bz == nil {
		return grant, false
	}
	k.cdc.MustUnmarshalBinaryBare(bz, &grant)
	return grant, true
}

func (k Keeper) update(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, updated types.Capability) {
	grant, found := k.getCapabilityGrant(ctx, k.getActorCapabilityKey(grantee, granter, updated.MsgType()))
	if !found {
		return
	}
	grant.Capability = updated
}

func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) sdk.Result {
	var res sdk.Result
	for _, msg := range msgs {
		signers := msg.GetSigners()
		if len(signers) != 1 {
			return sdk.ErrUnknownRequest("can only dispatch a delegated msg with 1 signer").Result()
		}
		granter := signers[0]
		if !bytes.Equal(granter, grantee) {
			capability, _ := k.GetCapability(ctx, grantee, granter, msg)
			if capability == nil {
				return sdk.ErrUnauthorized("unauthorized").Result()
			}
			allow, updated, del := capability.Accept(msg, ctx.BlockHeader())
			if !allow {
				return sdk.ErrUnauthorized("unauthorized").Result()
			}
			if del {
				k.Revoke(ctx, grantee, granter, msg)
			} else if updated != nil {
				k.update(ctx, grantee, granter, updated)
			}
		}
		res = k.router.Route(msg.Route())(ctx, msg)
		if !res.IsOK() {
			return res
		}
	}

	return sdk.Result{}
}

func (k Keeper) Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, capability types.Capability, expiration time.Time) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(types.CapabilityGrant{Capability: capability, Expiration: expiration})
	actor := k.getActorCapabilityKey(grantee, granter, capability.MsgType())
	store.Set(actor, bz)
}

func (k Keeper) Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(k.getActorCapabilityKey(grantee, granter, msgType))
}

func (k Keeper) GetCapability(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) (cap types.Capability, expiration time.Time) {
	grant, found := k.getCapabilityGrant(ctx, k.getActorCapabilityKey(grantee, granter, msgType))
	if !found {
		return nil, time.Time{}
	}
	if !grant.Expiration.IsZero() && grant.Expiration.Before(ctx.BlockHeader().Time) {
		k.Revoke(ctx, grantee, granter, msgType)
		return nil, time.Time{}
	}
	return grant.Capability, grant.Expiration
}
