package stake

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
	"github.com/pkg/errors"
)

// InitGenesis sets the pool and parameters for the provided keeper and
// initializes the IntraTxCounter. For each validator in data, it sets that
// validator in the keeper along with manually setting the indexes. In
// addition, it also sets any delegations found in data. Finally, it updates
// the bonded validators.
// Returns final validator set after applying all declaration and delegations
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) (res []abci.ValidatorUpdate, err error) {

	// We need to pretend to be "n blocks before genesis", where "n" is the validator update delay,
	// so that e.g. slashing periods are correctly initialized for the validator set
	// e.g. with a one-block offset - the first TM block is at height 0, so state updates applied from genesis.json are in block -1.
	ctx = ctx.WithBlockHeight(-types.ValidatorUpdateDelay)

	keeper.SetPool(ctx, data.Pool)
	keeper.SetParams(ctx, data.Params)

	for i, validator := range data.Validators {
		validator.BondIntraTxCounter = int16(i) // set the intra-tx counter to the order the validators are presented
		keeper.SetValidator(ctx, validator)

		if validator.Tokens.IsZero() {
			return res, errors.Errorf("genesis validator cannot have zero pool shares, validator: %v", validator)
		}
		if validator.DelegatorShares.IsZero() {
			return res, errors.Errorf("genesis validator cannot have zero delegator shares, validator: %v", validator)
		}

		// Manually set indices for the first time
		keeper.SetValidatorByConsAddr(ctx, validator)
		keeper.SetValidatorByPowerIndex(ctx, validator, data.Pool)
		keeper.OnValidatorCreated(ctx, validator.OperatorAddr)
	}

	for _, delegation := range data.Bonds {
		keeper.SetDelegation(ctx, delegation)
		keeper.OnDelegationCreated(ctx, delegation.DelegatorAddr, delegation.ValidatorAddr)
	}

	res = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	return
}

// WriteGenesis returns a GenesisState for a given context and keeper. The
// GenesisState will contain the pool, params, validators, and bonds found in
// the keeper.
func WriteGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	pool := keeper.GetPool(ctx)
	params := keeper.GetParams(ctx)
	validators := keeper.GetAllValidators(ctx)
	bonds := keeper.GetAllDelegations(ctx)

	return types.GenesisState{
		Pool:       pool,
		Params:     params,
		Validators: validators,
		Bonds:      bonds,
	}
}

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper Keeper) (vals []tmtypes.GenesisValidator) {
	keeper.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) (stop bool) {
		vals = append(vals, tmtypes.GenesisValidator{
			PubKey: validator.GetConsPubKey(),
			Power:  validator.GetPower().RoundInt64(),
			Name:   validator.GetMoniker(),
		})

		return false
	})

	return
}

// ValidateGenesis validates the provided staking genesis state to ensure the
// expected invariants holds. (i.e. params in correct bounds, no duplicate validators)
func ValidateGenesis(data types.GenesisState) error {
	err := validateGenesisStateValidators(data.Validators)
	if err != nil {
		return err
	}
	err = validateParams(data.Params)
	if err != nil {
		return err
	}

	return nil
}

func validateParams(params types.Params) error {
	if params.BondDenom == "" {
		return fmt.Errorf("staking parameter BondDenom can't be an empty string")
	}
	return nil
}

func validateGenesisStateValidators(validators []types.Validator) (err error) {
	addrMap := make(map[string]bool, len(validators))
	for i := 0; i < len(validators); i++ {
		val := validators[i]
		strKey := string(val.ConsPubKey.Bytes())
		if _, ok := addrMap[strKey]; ok {
			return fmt.Errorf("duplicate validator in genesis state: moniker %v, Address %v", val.Description.Moniker, val.ConsAddress())
		}
		if val.Jailed && val.Status == sdk.Bonded {
			return fmt.Errorf("validator is bonded and jailed in genesis state: moniker %v, Address %v", val.Description.Moniker, val.ConsAddress())
		}
		if val.Tokens.IsZero() {
			return fmt.Errorf("genesis validator cannot have zero pool shares, validator: %v", val)
		}
		if val.DelegatorShares.IsZero() {
			return fmt.Errorf("genesis validator cannot have zero delegator shares, validator: %v", val)
		}
		addrMap[strKey] = true
	}
	return
}
