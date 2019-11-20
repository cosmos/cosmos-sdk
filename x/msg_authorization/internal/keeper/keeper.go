package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
	"time"
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

func (k Keeper) DispatchActions(ctx sdk.Context, grantee sdk.AccAddress, msgs []sdk.Msg) sdk.Result {
	//TODO
	return sdk.Result{}
}

func (k Keeper) Grant(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, capability types.Capability, expiration time.Time) {
	//TODO
}

func (k Keeper) Revoke(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg)  {
	//TODO
}

func (k Keeper) GetCapability(ctx sdk.Context, grantee sdk.AccAddress, granter sdk.AccAddress, msgType sdk.Msg) (cap types.Capability, expiration time.Time)  {
	//TODO
	return nil, time.Now()
}
