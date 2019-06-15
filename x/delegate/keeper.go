package delegate

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmos "github.com/cosmos/cosmos-sdk/types"
)

type keeper struct {
	storeKey cosmos.StoreKey
	cdc      *codec.Codec
}

type capabilityGrant struct {
	// all the actors that delegated this capability to the actor
	// the capability should be cleared if root is false and this array is cleared
	delegatedBy []cosmos.AccAddress `json:"delegated_by"`

	// whenever this capability is undelegated or revoked, these delegations
	// need to be cleared recursively
	delegatedTo []cosmos.AccAddress `json:"delegated_to"`

	capability Capability `json:"capability"`
}

func NewKeeper(storeKey cosmos.StoreKey, cdc *codec.Codec) Keeper {
	return &keeper{storeKey: storeKey, cdc: cdc}
}

func ActorCapabilityKey(capability Capability, actor cosmos.AccAddress) []byte {
	return []byte(fmt.Sprintf("c/%s/%x", capability.CapabilityKey(), actor))
}

func (k keeper) getCapabilityGrant(ctx cosmos.Context, actor cosmos.AccAddress, capability Capability) (grant capabilityGrant, found bool) {
	//if bytes.Equal(actor, capability.RootAccount()) {
	//	return capabilityGrant{root:true}, true
	//}
	//store := ctx.KVStore(k.storeKey)
	//bz := store.Get(ActorCapabilityKey(capability, actor))
	//if bz == nil {
	//	return grant, false
	//}
	//k.cdc.MustUnmarshalBinaryBare(bz, &grant)
	//return grant, true
	panic("TODO")
}

func (k keeper) Delegate(ctx cosmos.Context, grantor cosmos.AccAddress, grantee cosmos.AccAddress, capability ActorCapability) bool {
	//store := ctx.KVStore(k.storeKey)
	//grantorGrant, found := k.getCapabilityGrant(ctx, grantor, capability)
	//if !found {
	//	return false
	//}
	//if !bytes.Equal(grantor, actor) {
	//
	//}
	//grantorGrant.delegatedTo = append(grantorGrant.delegatedTo, grantee)
	//store.Set(ActorCapabilityKey(capability, grantor), k.cdc.MustMarshalBinaryBare(grantorGrant))
	//granteeGrant, _ := k.getCapabilityGrant(ctx, grantee, capability)
	//granteeGrant.delegatedBy = append(granteeGrant.delegatedBy, grantor)
	//store.Set(ActorCapabilityKey(capability, grantee), k.cdc.MustMarshalBinaryBare(granteeGrant))
	//return true
	panic("TODO")
}

func (k keeper) Undelegate(ctx cosmos.Context, grantor cosmos.AccAddress, grantee cosmos.AccAddress, capability ActorCapability) {
	panic("implement me")
}

func (k keeper) HasCapability(ctx cosmos.Context, actor cosmos.AccAddress, capability ActorCapability) bool {
	//grant, found := k.getCapabilityGrant(ctx, actor, capability)
	//if !found {
	//	return false
	//}
	//if grant.root {
	//	return true
	//}
	//return grant.
	panic("TODO")
}
