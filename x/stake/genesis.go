package stake

import (
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	keeper.SetPool(ctx, data.Pool)
	keeper.SetNewParams(ctx, data.Params)
	keeper.InitIntraTxCounter(ctx)
	for _, validator := range data.Validators {

		// set validator
		keeper.SetValidator(ctx, validator)

		// manually set indexes for the first time
		keeper.SetValidatorByPubKeyIndex(ctx, validator)
		keeper.SetValidatorByPowerIndex(ctx, validator, data.Pool)
		if validator.Status() == sdk.Bonded {
			keeper.SetValidatorBondedIndex(ctx, validator)
		}
	}
	for _, bond := range data.Bonds {
		keeper.SetDelegation(ctx, bond)
	}
	keeper.UpdateBondedValidatorsFull(ctx)
}

// WriteGenesis - output genesis parameters
func WriteGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	pool := keeper.GetPool(ctx)
	params := keeper.GetParams(ctx)
	validators := keeper.GetAllValidators(ctx)
	bonds := keeper.GetAllDelegations(ctx)
	return types.GenesisState{
		pool,
		params,
		validators,
		bonds,
	}
}

// WriteValidators - output current validator set
func WriteValidators(ctx sdk.Context, keeper Keeper) (vals []tmtypes.GenesisValidator) {
	keeper.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
		vals = append(vals, tmtypes.GenesisValidator{
			PubKey: validator.GetPubKey(),
			Power:  validator.GetPower().RoundInt64(),
			Name:   validator.GetMoniker(),
		})
		return false
	})
	return
}
