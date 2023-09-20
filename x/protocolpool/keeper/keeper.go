package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"cosmossdk.io/x/protocolpool/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	storeService storetypes.KVStoreService
	authKeeper   types.AccountKeeper
	bankKeeper   types.BankKeeper

	authority string
}

func NewKeeper(cdc codec.BinaryCodec, storeService storetypes.KVStoreService,
	ak types.AccountKeeper, bk types.BankKeeper, authority string,
) Keeper {
	// ensure pool module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}
	return Keeper{
		storeService: storeService,
		authKeeper:   ak,
		bankKeeper:   bk,
		authority:    authority,
	}
}

// GetAuthority returns the x/protocolpool module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With(log.ModuleKey, "x/"+types.ModuleName)
}
