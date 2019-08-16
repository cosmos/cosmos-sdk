package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation parameter constants
const (
	UnbondingTime = "unbonding_time"
	MaxValidators = "max_validators"
)

// GenUnbondingTime randomized UnbondingTime
func GenUnbondingTime(cdc *codec.Codec, r *rand.Rand) (ubdTime time.Duration) {
	return time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
}

// GenMaxValidators randomized MaxValidators
func GenMaxValidators(cdc *codec.Codec, r *rand.Rand) (maxValidators uint16) {
	return uint16(r.Intn(250) + 1)
}

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(input *module.GeneratorInput) {

	var (
		validators  []types.Validator
		delegations []types.Delegation
	)

	// params
	// TODO: this could result in a bug if the staking module generator is not called
	// before the slashing generator !!! 
	input.UnbondTime = GenUnbondingTime(input.Cdc, input.R)
	maxValidators := GenMaxValidators(input.Cdc, input.R)
	params := types.NewParams(input.UnbondTime, maxValidators, 7, sdk.DefaultBondDenom)

	// validators & delegations
	valAddrs := make([]sdk.ValAddress, input.NumBonded)
	for i := 0; i < int(input.NumBonded); i++ {
		valAddr := sdk.ValAddress(input.Accounts[i].Address)
		valAddrs[i] = valAddr

		validator := types.NewValidator(valAddr, input.Accounts[i].PubKey, types.Description{})
		validator.Tokens = sdk.NewInt(input.InitialStake)
		validator.DelegatorShares = sdk.NewDec(input.InitialStake)
		delegation := types.NewDelegation(input.Accounts[i].Address, valAddr, sdk.NewDec(input.InitialStake))
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis := types.NewGenesisState(params, validators, delegations)

	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, stakingGenesis.Params))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(stakingGenesis)
}
