package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	authKeeper group.AccountKeeper
	bankKeeper group.BankKeeper
	server     serverImpl
}

// NewKeeper creates a new group keeper
func NewKeeper(storeKey storetypes.StoreKey, cdc codec.BinaryCodec, ak group.AccountKeeper, bk group.BankKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		authKeeper: ak,
		bankKeeper: bk,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
}
