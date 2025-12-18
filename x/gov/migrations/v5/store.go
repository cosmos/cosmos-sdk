package v5

import (
	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

var (
	// ParamsKey is the key of x/gov params
	ParamsKey = []byte{0x30}
	// ConstitutionKey is the key of x/gov constitution
	ConstitutionKey                          = collections.NewPrefix(64)
	ParticipationEMAKey                      = collections.NewPrefix(80)
	ConstitutionAmendmentParticipationEMAKey = collections.NewPrefix(96)
	LawParticipationEMAKey                   = collections.NewPrefix(112)
)

// MigrateStore performs in-place store migrations from v4 (v0.47) to v5 (v0.50). The
// migration includes:
//
// Addition of the new proposal expedited parameters that are set to 0 by default.
// Set of default chain constitution.
func MigrateStore(
	ctx sdk.Context,
	storeService corestoretypes.KVStoreService,
	cdc codec.BinaryCodec,
	constitutionItem collections.Item[string],
	participationEMAItem collections.Item[math.LegacyDec],
	constitutionAmendmentParticipationEMAItem collections.Item[math.LegacyDec],
	lawParticipationEMAItem collections.Item[math.LegacyDec],
) error {
	store := storeService.OpenKVStore(ctx)
	paramsBz, err := store.Get(ParamsKey)
	if err != nil {
		return err
	}

	var params govv1.Params
	err = cdc.Unmarshal(paramsBz, &params)
	if err != nil {
		return err
	}

	defaultParams := govv1.DefaultParams()
	params.ProposalCancelRatio = defaultParams.ProposalCancelRatio
	params.ProposalCancelDest = defaultParams.ProposalCancelDest
	params.MinDepositRatio = defaultParams.MinDepositRatio
	params.MinDepositThrottler = defaultParams.MinDepositThrottler
	params.MinDepositThrottler.FloorValue[0].Denom = sdk.DefaultBondDenom
	params.MinInitialDepositThrottler = defaultParams.MinInitialDepositThrottler
	params.MinInitialDepositThrottler.FloorValue[0].Denom = sdk.DefaultBondDenom
	params.BurnDepositNoThreshold = defaultParams.BurnDepositNoThreshold
	params.QuorumRange = defaultParams.QuorumRange
	params.ConstitutionAmendmentQuorumRange = defaultParams.ConstitutionAmendmentQuorumRange
	params.LawQuorumRange = defaultParams.LawQuorumRange
	params.GovernorStatusChangePeriod = defaultParams.GovernorStatusChangePeriod
	params.MinGovernorSelfDelegation = defaultParams.MinGovernorSelfDelegation

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	if err := store.Set(ParamsKey, bz); err != nil {
		return err
	}

	// Set other gov params
	initParticipationEma := math.LegacyNewDecWithPrec(12, 2)
	if ok, err := participationEMAItem.Has(ctx); !ok || err != nil {
		if err := participationEMAItem.Set(ctx, initParticipationEma); err != nil {
			return err
		}
	}

	if ok, err := constitutionAmendmentParticipationEMAItem.Has(ctx); !ok || err != nil {
		if err := constitutionAmendmentParticipationEMAItem.Set(ctx, initParticipationEma); err != nil {
			return err
		}
	}

	if ok, err := lawParticipationEMAItem.Has(ctx); !ok || err != nil {
		if err := lawParticipationEMAItem.Set(ctx, initParticipationEma); err != nil {
			return err
		}
	}

	// Set the default constitution if it is not set
	if ok, err := constitutionItem.Has(ctx); !ok || err != nil {
		if err := constitutionItem.Set(ctx, "This chain has no constitution."); err != nil {
			return err
		}
	}

	return nil
}
