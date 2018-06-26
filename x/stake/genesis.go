package stake

import (
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenesisState - all staking state that must be provided at genesis
type GenesisState struct {
	Pool       Pool         `json:"pool"`
	Params     Params       `json:"params"`
	Validators []Validator  `json:"validators"`
	Bonds      []Delegation `json:"bonds"`
}

func NewGenesisState(pool Pool, params Params, validators []Validator, bonds []Delegation) GenesisState {
	return GenesisState{
		Pool:       pool,
		Params:     params,
		Validators: validators,
		Bonds:      bonds,
	}
}

// get raw genesis raw message for testing
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Pool:   InitialPool(),
		Params: DefaultParams(),
	}
}

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	store := ctx.KVStore(k.storeKey)
	k.setPool(ctx, data.Pool)
	k.setNewParams(ctx, data.Params)
	for _, validator := range data.Validators {

		// set validator
		k.setValidator(ctx, validator)

		// manually set indexes for the first time
		k.setValidatorByPubKeyIndex(ctx, validator)
		k.setValidatorByPowerIndex(ctx, validator, data.Pool)
		if validator.Status() == sdk.Bonded {
			store.Set(GetValidatorsBondedKey(validator.PubKey), validator.Owner)
		}
	}
	for _, bond := range data.Bonds {
		k.setDelegation(ctx, bond)
	}
	k.updateBondedValidatorsFull(ctx, store)
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)
	validators := k.getAllValidators(ctx)
	bonds := k.getAllDelegations(ctx)
	return GenesisState{
		pool,
		params,
		validators,
		bonds,
	}
}

// WriteValidators - output current validator set
func WriteValidators(ctx sdk.Context, k Keeper) (vals []tmtypes.GenesisValidator) {
	k.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
		vals = append(vals, tmtypes.GenesisValidator{
			PubKey: validator.GetPubKey(),
			Power:  validator.GetPower().Evaluate(),
			Name:   validator.GetMoniker(),
		})
		return false
	})
	return
}
