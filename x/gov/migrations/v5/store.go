package v5

import (
	corestoretypes "cosmossdk.io/core/store"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// MigrateStore performs in-place store migrations from v4 (v0.47) to v5 (v0.50). The
// migration includes:
//
// Addition of the new proposal expedited parameters that are set to 0 by default.
func MigrateStore(ctx sdk.Context, storeService corestoretypes.KVStoreService, cdc codec.BinaryCodec) error {
	store := storeService.OpenKVStore(ctx)
	paramsBz, err := store.Get(v4.ParamsKey)
	if err != nil {
		return err
	}

	var params govv1.Params
	err = cdc.Unmarshal(paramsBz, &params)
	if err != nil {
		return err
	}

	defaultParams := govv1.DefaultParams()
	params.ExpeditedMinDeposit = defaultParams.ExpeditedMinDeposit
	params.ExpeditedVotingPeriod = defaultParams.ExpeditedVotingPeriod
	params.ExpeditedThreshold = defaultParams.ExpeditedThreshold
	params.ProposalCancelRatio = defaultParams.ProposalCancelRatio
	params.ProposalCancelDest = defaultParams.ProposalCancelDest

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	return store.Set(v4.ParamsKey, bz)
}
