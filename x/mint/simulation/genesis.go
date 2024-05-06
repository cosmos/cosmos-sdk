package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"cosmossdk.io/math"
	"cosmossdk.io/x/mint/types"

	"github.com/cosmos/cosmos-sdk/types/module"
)

// Simulation parameter constants
const (
	Inflation                            = "inflation"
	InflationRateChange                  = "inflation_rate_change"
	InflationMax                         = "inflation_max"
	InflationMin                         = "inflation_min"
	GoalBonded                           = "goal_bonded"
	ReductionFactor                      = "reduction_factor"
	ReductionStartedEpoch                = "reduction_started_epoch"
	ReductionPeriodInEpochs              = "reduction_period_in_epochs"
	MintingRewardsDistributionStartEpoch = "minting_rewards_distribution_start_epoch"
	EpochProvisions                      = "epoch_provisions"

	epochIdentifier = "day"
	maxInt64        = int(^uint(0) >> 1)
)

// GenInflation randomized Inflation
func GenInflation(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationRateChange randomized InflationRateChange
func GenInflationRateChange(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(r.Intn(99)), 2)
}

// GenInflationMax randomized InflationMax
func GenInflationMax(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(20, 2)
}

// GenInflationMin randomized InflationMin
func GenInflationMin(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(7, 2)
}

// GenGoalBonded randomized GoalBonded
func GenGoalBonded(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(67, 2)
}

func GenReductionFactor(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(r.Intn(10)), 1)
}

func GenReductionStartedEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func GenMintingRewardsDistributionStartEpoch(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func GenReductionPeriodInEpochs(r *rand.Rand) int64 {
	return int64(r.Intn(maxInt64))
}

func GenEpochProvisions(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDec(int64(r.Intn(maxInt64)))
}

// RandomizedGenState generates a random GenesisState for mint
func RandomizedGenState(simState *module.SimulationState) {
	// minter
	var inflation math.LegacyDec
	simState.AppParams.GetOrGenerate(Inflation, &inflation, simState.Rand, func(r *rand.Rand) { inflation = GenInflation(r) })

	// params
	var inflationRateChange math.LegacyDec
	simState.AppParams.GetOrGenerate(InflationRateChange, &inflationRateChange, simState.Rand, func(r *rand.Rand) { inflationRateChange = GenInflationRateChange(r) })

	var inflationMax math.LegacyDec
	simState.AppParams.GetOrGenerate(InflationMax, &inflationMax, simState.Rand, func(r *rand.Rand) { inflationMax = GenInflationMax(r) })

	var inflationMin math.LegacyDec
	simState.AppParams.GetOrGenerate(InflationMin, &inflationMin, simState.Rand, func(r *rand.Rand) { inflationMin = GenInflationMin(r) })

	var goalBonded math.LegacyDec
	simState.AppParams.GetOrGenerate(GoalBonded, &goalBonded, simState.Rand, func(r *rand.Rand) { goalBonded = GenGoalBonded(r) })

	var reductionFactor math.LegacyDec
	simState.AppParams.GetOrGenerate(ReductionFactor, &reductionFactor, simState.Rand, func(r *rand.Rand) { reductionFactor = GenReductionFactor(r) })

	var reductionStartedEpoch int64
	simState.AppParams.GetOrGenerate(ReductionStartedEpoch, &reductionStartedEpoch, simState.Rand, func(r *rand.Rand) { reductionStartedEpoch = GenReductionStartedEpoch(r) })

	var reductionPeriodInEpochs int64
	simState.AppParams.GetOrGenerate(ReductionPeriodInEpochs, &reductionPeriodInEpochs, simState.Rand, func(r *rand.Rand) { reductionPeriodInEpochs = GenReductionPeriodInEpochs(r) })

	var mintingRewardsDistributionStartEpoch int64
	simState.AppParams.GetOrGenerate(MintingRewardsDistributionStartEpoch, &mintingRewardsDistributionStartEpoch, simState.Rand, func(r *rand.Rand) { mintingRewardsDistributionStartEpoch = GenMintingRewardsDistributionStartEpoch(r) })

	var epochProvisions math.LegacyDec
	simState.AppParams.GetOrGenerate(EpochProvisions, &epochProvisions, simState.Rand, func(r *rand.Rand) { epochProvisions = GenEpochProvisions(r) })

	mintDenom := simState.BondDenom
	blocksPerYear := uint64(60 * 60 * 8766 / 5)
	params := types.NewParams(
		mintDenom,
		inflationRateChange,
		inflationMax,
		inflationMin,
		goalBonded,
		blocksPerYear,
		math.ZeroInt(),
		epochIdentifier,
		reductionPeriodInEpochs,
		reductionFactor,
		mintingRewardsDistributionStartEpoch,
		epochProvisions,
	)

	mintGenesis := types.NewGenesisState(types.InitialMinter(inflation), params, reductionStartedEpoch)

	bz, err := json.MarshalIndent(&mintGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated minting parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(mintGenesis)
}
