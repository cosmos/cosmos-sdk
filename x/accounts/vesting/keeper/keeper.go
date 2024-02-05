package keeper

import (
	"context"

	vestingtypes "cosmossdk.io/x/accounts/vesting/types"
)

type Keeper struct {
	bankKeeper     vestingtypes.BankKeeper
	accountsKeeper vestingtypes.AccountsKeeper
}

func (k Keeper) Init(ctx context.Context)

func (k Keeper) ExecuteMsgs(ctx context.Context)
