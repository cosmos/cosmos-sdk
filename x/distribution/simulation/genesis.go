package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// Simulation parameter constants
const (
	CommunityTax         = "community_tax"
	WithdrawEnabled      = "withdraw_enabled"
	NakamotoBonusEnabled = "nakamoto_bonus_enabled"
	NakamotoBonusStep    = "nakamoto_bonus_step"
	NakamotoBonusPeriod  = "nakamoto_bonus_enabled"
	NakamotoBonusMax     = "nakamoto_bonus_maximum"
	NakamotoBonusMin     = "nakamoto_bonus_minimum"
)

// GenCommunityTax randomized CommunityTax
func GenCommunityTax(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(1, 2).Add(math.LegacyNewDecWithPrec(int64(r.Intn(30)), 2))
}

// GenWithdrawEnabled returns a randomized WithdrawEnabled parameter.
func GenWithdrawEnabled(r *rand.Rand) bool {
	return r.Int63n(101) <= 95 // 95% chance of withdraws being enabled
}

// GenNakamotoBonusEnabled returns a randomized NakamotoBonusEnabled parameter.
func GenNakamotoBonusEnabled(r *rand.Rand) bool {
	return r.Int63n(101) <= 70 // 70% chance of nakamoto bonus being enabled
}

// GenNakamotoBonusStep returns a randomized NakamotoBonusStep parameter.
func GenNakamotoBonusStep(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(1, 2).Add(math.LegacyNewDecWithPrec(int64(r.Intn(10)), 2))
}

// GenNakamotoBonusPeriodEpochIdentifier returns a randomized NakamotoBonusPeriodEpochIdentifier parameter.
func GenNakamotoBonusPeriodEpochIdentifier(r *rand.Rand) string {
	// Choose randomly between common epoch identifiers
	identifiers := []string{"day", "week", "month"}
	return identifiers[r.Intn(len(identifiers))]
}

// RandomizedGenState generates a random GenesisState for distribution
func RandomizedGenState(simState *module.SimulationState) {
	var communityTax math.LegacyDec
	simState.AppParams.GetOrGenerate(CommunityTax, &communityTax, simState.Rand, func(r *rand.Rand) { communityTax = GenCommunityTax(r) })

	var withdrawEnabled bool
	simState.AppParams.GetOrGenerate(WithdrawEnabled, &withdrawEnabled, simState.Rand, func(r *rand.Rand) { withdrawEnabled = GenWithdrawEnabled(r) })

	var nakamotoBonusEnabled bool
	simState.AppParams.GetOrGenerate(NakamotoBonusEnabled, &nakamotoBonusEnabled, simState.Rand, func(r *rand.Rand) { nakamotoBonusEnabled = GenNakamotoBonusEnabled(r) })

	var nakamotoBonusStep math.LegacyDec
	simState.AppParams.GetOrGenerate(NakamotoBonusStep, &nakamotoBonusStep, simState.Rand, func(r *rand.Rand) { nakamotoBonusStep = GenNakamotoBonusStep(r) })

	var nakamotoBonusPeriodEpochIdentifier string
	simState.AppParams.GetOrGenerate(NakamotoBonusPeriod, &nakamotoBonusPeriodEpochIdentifier, simState.Rand, func(r *rand.Rand) { nakamotoBonusPeriodEpochIdentifier = GenNakamotoBonusPeriodEpochIdentifier(r) })

	var nakamotoBonusMin math.LegacyDec
	simState.AppParams.GetOrGenerate(NakamotoBonusMin, &nakamotoBonusMin, simState.Rand, func(r *rand.Rand) { nakamotoBonusMin = GenNakamotoBonusStep(r) })

	var nakamotoBonusMax math.LegacyDec
	simState.AppParams.GetOrGenerate(NakamotoBonusMax, &nakamotoBonusMax, simState.Rand, func(r *rand.Rand) { nakamotoBonusMax = GenNakamotoBonusStep(r) })

	distrGenesis := types.GenesisState{
		FeePool: types.InitialFeePool(),
		Params: types.Params{
			CommunityTax:        communityTax,
			WithdrawAddrEnabled: withdrawEnabled,
			NakamotoBonus: types.NakamotoBonus{
				Enabled:               nakamotoBonusEnabled,
				Step:                  nakamotoBonusStep,
				PeriodEpochIdentifier: nakamotoBonusPeriodEpochIdentifier,
				MinimumCoefficient:    nakamotoBonusMin,
				MaximumCoefficient:    nakamotoBonusMax,
			},
		},
	}

	bz, err := json.MarshalIndent(&distrGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated distribution parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&distrGenesis)
}
