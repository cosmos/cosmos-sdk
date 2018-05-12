package stake

import sdk "github.com/cosmos/cosmos-sdk/types"

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
		Pool:   initialPool(),
		Params: defaultParams(),
	}
}

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, k Keeper, data GenesisState) {
	k.setPool(ctx, data.Pool)
	k.setParams(ctx, data.Params)
	for _, validator := range data.Validators {
		k.setValidator(ctx, validator)
	}
	for _, bond := range data.Bonds {
		k.setDelegation(ctx, bond)
	}
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, k Keeper) GenesisState {
	pool := k.GetPool(ctx)
	params := k.GetParams(ctx)
	validators := k.GetValidators(ctx, 32767)
	bonds := k.getBonds(ctx, 32767)
	return GenesisState{
		pool,
		params,
		validators,
		bonds,
	}
}
