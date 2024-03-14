package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// BeginBlocker will persist the current header and validator set as a historical entry
// and prune the oldest entry based on the HistoricalEntries parameter
func (k *Keeper) BeginBlocker(ctx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	return k.TrackHistoricalInfo(ctx)
}

// EndBlocker called at every block, update validator set
func (k *Keeper) EndBlocker(ctx context.Context) ([]module.ValidatorUpdate, error) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	cometValidatorUpdates, err := k.BlockValidatorUpdates(ctx)
	if err != nil {
		return nil, err
	}
	validatorUpdates := make([]module.ValidatorUpdate, len(cometValidatorUpdates))
	for i, v := range cometValidatorUpdates {
		if ed25519 := v.PubKey.GetEd25519(); len(ed25519) > 0 {
			validatorUpdates[i] = module.ValidatorUpdate{
				PubKey:     ed25519,
				PubKeyType: "ed25519",
				Power:      v.Power,
			}
		} else if secp256k1 := v.PubKey.GetSecp256K1(); len(secp256k1) > 0 {
			validatorUpdates[i] = module.ValidatorUpdate{
				PubKey:     secp256k1,
				PubKeyType: "secp256k1",
				Power:      v.Power,
			}
		} else {
			return nil, fmt.Errorf("unexpected validator pubkey type: %T", v.PubKey)
		}
	}
	return validatorUpdates, nil
}
