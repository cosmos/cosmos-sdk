package v4

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// If expedited, the deposit to enter voting period will be
// increased to 5000 OSMO. The proposal will have 24 hours to achieve
// a two-thirds majority of all staked OSMO voting power voting YES.

var (
	minExpeditedDeposit = sdk.NewCoins(sdk.NewCoin("uosmo", sdk.NewInt(5000 * 1_000_000)))
	expeditedVotingPeriod = time.Duration(time.Hour * 24)
	expeditedThreshold = sdk.NewDec(2).Quo(sdk.NewDec(3))
)

// MigrateStore performs in-place store migrations for consensus version 4
// in the gov module.
// The migration includes:
//
// - Setting the expedited proposals params in the paramstore.
func MigrateStore(ctx sdk.Context, paramstore types.ParamSubspace) error {
	migrateParamsStore(ctx, paramstore)
	return nil
}

func migrateParamsStore(ctx sdk.Context, paramstore types.ParamSubspace) {
	var (
		depositParams types.DepositParams 
		votingParams types.VotingParams 
		tallyParams types.TallyParams
	)

	// Set depositParams
	paramstore.Get(ctx, types.ParamStoreKeyDepositParams, &depositParams)
	depositParams.MinExpeditedDeposit = minExpeditedDeposit
	paramstore.Set(ctx, types.ParamStoreKeyDepositParams, depositParams)

	// Set votingParams
	paramstore.Get(ctx, types.ParamStoreKeyVotingParams, &votingParams)
	votingParams.ExpeditedVotingPeriod = expeditedVotingPeriod
	paramstore.Set(ctx, types.ParamStoreKeyVotingParams, votingParams)

	// Set tallyParams
	paramstore.Get(ctx, types.ParamStoreKeyTallyParams, &tallyParams)
	tallyParams.ExpeditedThreshold = expeditedThreshold
	paramstore.Set(ctx, types.ParamStoreKeyTallyParams, tallyParams)
}
