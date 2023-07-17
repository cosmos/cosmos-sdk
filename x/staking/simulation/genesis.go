package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Simulation parameter constants
const (
	UnbondingTime             = "unbonding_time"
	MaxValidators             = "max_validators"
	HistoricalEntries         = "historical_entries"
	ValidatorBondFactor       = "validator_bond_factor"
	GlobalLiquidStakingCap    = "global_liquid_staking_cap"
	ValidatorLiquidStakingCap = "validator_liquid_staking_cap"
)

// genUnbondingTime returns randomized UnbondingTime
func genUnbondingTime(r *rand.Rand) (ubdTime time.Duration) {
	return time.Duration(simulation.RandIntBetween(r, 60, 60*60*24*3*2)) * time.Second
}

// genMaxValidators returns randomized MaxValidators
func genMaxValidators(r *rand.Rand) (maxValidators uint32) {
	return uint32(r.Intn(250) + 1)
}

// getHistEntries returns randomized HistoricalEntries between 0-100.
func getHistEntries(r *rand.Rand) uint32 {
	return uint32(r.Intn(int(types.DefaultHistoricalEntries + 1)))
}

// getGlobalLiquidStakingCap returns randomized GlobalLiquidStakingCap between 0-1.
func getGlobalLiquidStakingCap(r *rand.Rand) sdk.Dec {
	return simulation.RandomDecAmount(r, sdk.OneDec())
}

// getValidatorLiquidStakingCap returns randomized ValidatorLiquidStakingCap between 0-1.
func getValidatorLiquidStakingCap(r *rand.Rand) sdk.Dec {
	return simulation.RandomDecAmount(r, sdk.OneDec())
}

// getValidatorBondFactor returns randomized ValidatorBondCap between -1 and 300.
func getValidatorBondFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(simulation.RandIntBetween(r, -1, 300)))
}

// RandomizedGenState generates a random GenesisState for staking
func RandomizedGenState(simState *module.SimulationState) {
	// params
	var (
		unbondingTime             time.Duration
		maxValidators             uint32
		historicalEntries         uint32
		minCommissionRate         sdk.Dec
		validatorBondFactor       sdk.Dec
		globalLiquidStakingCap    sdk.Dec
		validatorLiquidStakingCap sdk.Dec
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, UnbondingTime, &unbondingTime, simState.Rand,
		func(r *rand.Rand) { unbondingTime = genUnbondingTime(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxValidators, &maxValidators, simState.Rand,
		func(r *rand.Rand) { maxValidators = genMaxValidators(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, HistoricalEntries, &historicalEntries, simState.Rand,
		func(r *rand.Rand) { historicalEntries = getHistEntries(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, ValidatorBondFactor, &validatorBondFactor, simState.Rand,
		func(r *rand.Rand) { validatorBondFactor = getValidatorBondFactor(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, GlobalLiquidStakingCap, &globalLiquidStakingCap, simState.Rand,
		func(r *rand.Rand) { globalLiquidStakingCap = getGlobalLiquidStakingCap(r) },
	)

	simState.AppParams.GetOrGenerate(
		simState.Cdc, ValidatorLiquidStakingCap, &validatorLiquidStakingCap, simState.Rand,
		func(r *rand.Rand) { validatorLiquidStakingCap = getValidatorLiquidStakingCap(r) },
	)

	// NOTE: the slashing module need to be defined after the staking module on the
	// NewSimulationManager constructor for this to work
	simState.UnbondTime = unbondingTime
	params := types.NewParams(
		simState.UnbondTime,
		maxValidators,
		7,
		historicalEntries,
		sdk.DefaultBondDenom,
		minCommissionRate,
		validatorBondFactor,
		globalLiquidStakingCap,
		validatorLiquidStakingCap,
	)

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

		validator, err := types.NewValidator(valAddr, simState.Accounts[i].ConsKey.PubKey(), types.Description{})
		if err != nil {
			panic(err)
		}
		validator.Tokens = simState.InitialStake
		validator.DelegatorShares = sdk.NewDecFromInt(simState.InitialStake)
		validator.Commission = commission

		delegation := types.NewDelegation(simState.Accounts[i].Address, valAddr, sdk.NewDecFromInt(simState.InitialStake), false)

		validators = append(validators, validator)
		delegations = append(delegations, delegation)
	}

	stakingGenesis := types.NewGenesisState(params, validators, delegations)

	bz, err := json.MarshalIndent(&stakingGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated staking parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(stakingGenesis)
}
