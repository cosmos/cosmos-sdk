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
func GenUnbondingTime(r *rand.Rand) (ubdTime time.Duration) {
	return time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
}

// GenMaxValidators randomized MaxValidators
func GenMaxValidators(r *rand.Rand) (maxValidators uint16) {
	return uint16(r.Intn(250) + 1)
}

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(simState *module.SimulationState) {
	// params
	var unbondTime time.Duration
	simState.AppParams.GetOrGenerate(
		simState.Cdc, UnbondingTime, &unbondTime, simState.Rand,
		func(r *rand.Rand) { unbondTime = GenUnbondingTime(r) },
	)

	var maxValidators uint16
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxValidators, &maxValidators, simState.Rand,
		func(r *rand.Rand) { maxValidators = GenMaxValidators(r) },
	)

	// NOTE: the slashing module need to be defined after the staking module on the
	// NewSimulationManager constructor for this to work
	simState.UnbondTime = unbondTime

	params := types.NewParams(simState.UnbondTime, maxValidators, 7, 3, sdk.DefaultBondDenom)

	// validators & delegations
	var (
		validators  []types.Validator
		delegations []types.Delegation
	)

	valAddrs := make([]sdk.ValAddress, simState.NumBonded)
	for i := 0; i < int(simState.NumBonded); i++ {
		valAddr := sdk.ValAddress(simState.Accounts[i].Address)
		valAddrs[i] = valAddr

		maxCommission := sdk.NewDecWithPrec(int64(simulation.RandIntBetween(simState.Rand, 1, 100)), 2)
		commission := types.NewCommission(
			simulation.RandomDecAmount(simState.Rand, maxCommission),
			maxCommission,
			simulation.RandomDecAmount(simState.Rand, maxCommission),
		)

		validator := types.NewValidator(valAddr, simState.Accounts[i].PubKey, types.Description{})
		validator.Tokens = sdk.NewInt(simState.InitialStake)
		validator.DelegatorShares = sdk.NewDec(simState.InitialStake)
		validator.Commission = commission

		delegation := types.NewDelegation(simState.Accounts[i].Address, valAddr, sdk.NewDec(simState.InitialStake))
		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis := types.NewGenesisState(params, validators, delegations)

	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", codec.MustMarshalJSONIndent(simState.Cdc, stakingGenesis.Params))
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(stakingGenesis)
}
