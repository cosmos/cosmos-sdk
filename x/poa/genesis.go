package poa

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/poa/exported"
	"github.com/cosmos/cosmos-sdk/x/poa/internal/types"
)

// InitGenesis sets the parameters for the provided keeper. For each
// validator in data, it sets that validator in the keeper along with manually
// setting the indexes. Finally, it updates the bonded validators.
// Returns final validator set after applying all declaration
func InitGenesis(ctx sdk.Context, keeper Keeper, accountKeeper AccountKeeper,
	supplyKeeper SupplyKeeper, data GenesisState) (res []abci.ValidatorUpdate) {

	// We need to pretend to be "n blocks before genesis", where "n" is the
	// validator update delay, so that e.g. slashing periods are correctly
	// initialized for the validator set e.g. with a one-block offset - the
	// first TM block is at height 1, so state updates applied from
	// genesis.json are in block 0.
	ctx = ctx.WithBlockHeight(1 - sdk.ValidatorUpdateDelay)

	keeper.SetParams(ctx, data.Params)
	keeper.SetLastTotalPower(ctx, data.LastTotalPower)

	for _, validator := range data.Validators {
		keeper.SetValidator(ctx, validator)

		// Manually set indices for the first time
		keeper.SetValidatorByConsAddr(ctx, validator)
		keeper.SetValidatorByPowerIndex(ctx, validator)

		// Call the creation hook if not exported
		if !data.Exported {
			keeper.AfterValidatorCreated(ctx, validator.OperatorAddress)
		}

		// update timeslice if necessary
		if validator.IsUnbonding() {
			keeper.InsertValidatorQueue(ctx, validator)
		}
	}

	// don't need to run Tendermint updates if we exported
	if data.Exported {
		for _, lv := range data.LastValidatorPowers {
			keeper.SetLastValidatorPower(ctx, lv.Address, lv.Power)
			validator, found := keeper.GetValidator(ctx, lv.Address)
			if !found {
				panic(fmt.Sprintf("validator %s not found", lv.Address))
			}
			update := validator.ABCIValidatorUpdate()
			update.Power = lv.Power // keep the next-val-set offset, use the last power for the first block
			res = append(res, update)
		}
	} else {
		res = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	}

	return res
}

// ExportGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
	params := keeper.GetParams(ctx)
	lastTotalPower := keeper.GetLastTotalPower(ctx)
	validators := keeper.GetAllValidators(ctx)

	var lastValidatorPowers []types.LastValidatorPower
	keeper.IterateLastValidatorPowers(ctx, func(addr sdk.ValAddress, power int64) (stop bool) {
		lastValidatorPowers = append(lastValidatorPowers, types.LastValidatorPower{Address: addr, Power: power})
		return false
	})

	return GenesisState{
		Params:              params,
		LastTotalPower:      lastTotalPower,
		LastValidatorPowers: lastValidatorPowers,
		Validators:          validators,
		Exported:            true,
	}
}

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper Keeper) (vals []tmtypes.GenesisValidator) {
	keeper.IterateLastValidators(ctx, func(_ int64, validator exported.ValidatorI) (stop bool) {
		vals = append(vals, tmtypes.GenesisValidator{
			PubKey: validator.GetConsPubKey(),
			Power:  validator.GetConsensusPower(),
			Name:   validator.GetMoniker(),
		})

		return false
	})

	return
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func ValidateGenesis(data GenesisState) error {
	err := validateGenesisStateValidators(data.Validators)
	if err != nil {
		return err
	}
	err = data.Params.Validate()
	if err != nil {
		return err
	}

	return nil
}

func validateGenesisStateValidators(validators []Validator) (err error) {
	addrMap := make(map[string]bool, len(validators))
	for i := 0; i < len(validators); i++ {
		val := validators[i]
		strKey := string(val.ConsPubKey.Bytes())
		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, address %v", val.Description.Moniker, val.ConsAddress())
		}
		if val.Jailed && val.IsBonded() {
			return fmt.Errorf("validator is bonded and jailed in genesis state: moniker %v, address %v", val.Description.Moniker, val.ConsAddress())
		}
		addrMap[strKey] = true
	}
	return
}
