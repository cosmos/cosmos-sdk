package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// Simulation parameter constants
const (
	MinDeposit            = "min_deposit"
	ExpeditedMinDeposit   = "expedited_min_deposit"
	DepositPeriod         = "deposit_period"
	MinInitialRatio       = "min_initial_ratio"
	VotingPeriod          = "voting_period"
	ExpeditedVotingPeriod = "expedited_voting_period"
	Quorum                = "quorum"
	Threshold             = "threshold"
	ExpeditedThreshold    = "expedited_threshold"
	Veto                  = "veto"
	ProposalCancelRate    = "proposal_cancel_rate"

	// ExpeditedThreshold must be at least as large as the regular Threshold
	// Therefore, we use this break out point in randomization.
	tallyNonExpeditedMax = 500

	// Similarly, expedited voting period must be strictly less than the regular
	// voting period to be valid. Therefore, we use this break out point in randomization.
	expeditedMaxVotingPeriod = 60 * 60 * 24 * 2
)

// GenDepositPeriod returns randomized DepositPeriod
func GenDepositPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenMinDeposit returns randomized MinDeposit
func GenMinDeposit(r *rand.Rand, bondDenom string) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(bondDenom, int64(simulation.RandIntBetween(r, 1, 1e3/2))))
}

// GenExpeditedMinDeposit returns randomized ExpeditedMinDeposit
// It is always greater than GenMinDeposit
func GenExpeditedMinDeposit(r *rand.Rand, bondDenom string) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(bondDenom, int64(simulation.RandIntBetween(r, 1e3/2, 1e3))))
}

// GenDepositMinInitialRatio returns randomized DepositMinInitialRatio
func GenDepositMinInitialDepositRatio(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(sdk.NewDec(100))
}

// GenProposalCancelRate returns randomized ProposalCancelRate
func GenProposalCancelRate(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(sdk.NewDec(100))
}

// GenVotingPeriod returns randomized VotingPeriod
func GenVotingPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, expeditedMaxVotingPeriod, 2*expeditedMaxVotingPeriod)) * time.Second
}

// GenExpeditedVotingPeriod randomized ExpeditedVotingPeriod
func GenExpeditedVotingPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, expeditedMaxVotingPeriod)) * time.Second
}

// GenQuorum returns randomized Quorum
func GenQuorum(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 334, 500)), 3)
}

// GenThreshold returns randomized Threshold
func GenThreshold(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 450, tallyNonExpeditedMax+1)), 3)
}

// GenExpeditedThreshold randomized ExpeditedThreshold
func GenExpeditedThreshold(r *rand.Rand) sdk.Dec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, tallyNonExpeditedMax, 550)), 3)
}

// GenVeto returns randomized Veto
func GenVeto(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 250, 334)), 3)
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	startingProposalID := uint64(simState.Rand.Intn(100))

	var minDeposit sdk.Coins
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MinDeposit, &minDeposit, simState.Rand,
		func(r *rand.Rand) { minDeposit = GenMinDeposit(r, simState.BondDenom) },
	)

	var expeditedMinDeposit sdk.Coins
	simState.AppParams.GetOrGenerate(
		simState.Cdc, ExpeditedMinDeposit, &expeditedMinDeposit, simState.Rand,
		func(r *rand.Rand) { expeditedMinDeposit = GenExpeditedMinDeposit(r, simState.BondDenom) },
	)

	var depositPeriod time.Duration
	simState.AppParams.GetOrGenerate(
		simState.Cdc, DepositPeriod, &depositPeriod, simState.Rand,
		func(r *rand.Rand) { depositPeriod = GenDepositPeriod(r) },
	)

	var minInitialDepositRatio sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, MinInitialRatio, &minInitialDepositRatio, simState.Rand,
		func(r *rand.Rand) { minInitialDepositRatio = GenDepositMinInitialDepositRatio(r) },
	)

	var proposalCancelRate sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, ProposalCancelRate, &proposalCancelRate, simState.Rand,
		func(r *rand.Rand) { proposalCancelRate = GenProposalCancelRate(r) },
	)

	var votingPeriod time.Duration
	simState.AppParams.GetOrGenerate(
		simState.Cdc, VotingPeriod, &votingPeriod, simState.Rand,
		func(r *rand.Rand) { votingPeriod = GenVotingPeriod(r) },
	)

	var expeditedVotingPeriod time.Duration
	simState.AppParams.GetOrGenerate(
		simState.Cdc, ExpeditedVotingPeriod, &expeditedVotingPeriod, simState.Rand,
		func(r *rand.Rand) { expeditedVotingPeriod = GenExpeditedVotingPeriod(r) },
	)

	var quorum sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Quorum, &quorum, simState.Rand,
		func(r *rand.Rand) { quorum = GenQuorum(r) },
	)

	var threshold sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Threshold, &threshold, simState.Rand,
		func(r *rand.Rand) { threshold = GenThreshold(r) },
	)

	var expitedVotingThreshold sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, ExpeditedThreshold, &expitedVotingThreshold, simState.Rand,
		func(r *rand.Rand) { expitedVotingThreshold = GenExpeditedThreshold(r) },
	)

	var veto sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, Veto, &veto, simState.Rand,
		func(r *rand.Rand) { veto = GenVeto(r) },
	)

	govGenesis := v1.NewGenesisState(
		startingProposalID,
		v1.NewParams(minDeposit, expeditedMinDeposit, depositPeriod, votingPeriod, expeditedVotingPeriod, quorum.String(), threshold.String(), expitedVotingThreshold.String(), veto.String(), minInitialDepositRatio.String(), proposalCancelRate.String(), "", simState.Rand.Intn(2) == 0, simState.Rand.Intn(2) == 0, simState.Rand.Intn(2) == 0),
	)

	bz, err := json.MarshalIndent(&govGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(govGenesis)
}
