package genutil

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// expected staking keeper
type StakingKeeper interface {
	ApplyAndReturnValidatorSetUpdates(sdk.Context) (updates []abci.ValidatorUpdate)
}

type AccountKeeper interface {
	NewAccount(sdk.Context, auth.Account)
	SetAccount(sdk.Context, auth.Account)
	IterateAccounts(ctx sdk.Context, process func(Account) (stop bool))
}
