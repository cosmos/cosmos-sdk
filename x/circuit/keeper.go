package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper stores keytable-initialized params.Subspace for circuit breaking
type Keeper struct {
	space params.Subspace
}

// NewKeeper constructs new keeper
func NewKeeper(space params.Subspace) Keeper {
	return Keeper{
		space: space.WithKeyTable(ParamKeyTable()),
	}
}

func (k Keeper) BreakRoute(ctx sdk.Context, msg sdk.Msg) {
	k.space.SetWithSubkey(ctx, MsgRouteKey, []byte(msg.Route()), true)
}

func (k Keeper) BreakType(ctx sdk.Context, msg sdk.Msg) {
	k.space.SetWithSubkey(ctx, MsgTypeKey, []byte(msg.Type()), true)
}

func (k Keeper) RecoverRoute(ctx sdk.Context, msg sdk.Msg) {
	k.space.SetWithSubkey(ctx, MsgRouteKey, []byte(msg.Route()), false)
}

func (k Keeper) RecoverType(ctx sdk.Context, msg sdk.Msg) {
	k.space.SetWithSubkey(ctx, MsgTypeKey, []byte(msg.Type()), false)
}

func (k Keeper) CheckMsgBreak(ctx sdk.Context, msg sdk.Msg) (brk bool) {
	k.space.GetWithSubkeyIfExists(ctx, MsgRouteKey, []byte(msg.Route()), &brk)
	if brk {
		return
	}
	k.space.GetWithSubkeyIfExists(ctx, MsgTypeKey, []byte(msg.Type()), &brk)
	return
}
