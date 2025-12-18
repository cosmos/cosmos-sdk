package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// Simulation parameter constants
const (
	DepositParamsMinDeposit    = "deposit_params_min_deposit"
	DepositParamsDepositPeriod = "deposit_params_deposit_period"
	// DepositMinInitialRatio                                  = "deposit_params_min_initial_ratio"
	VotingParamsVotingPeriod                                = "voting_params_voting_period"
	TallyParamsThreshold                                    = "tally_params_threshold"
	TallyParamsConstitutionAmendmentThreshold               = "tally_params_constitution_amendment_threshold"
	TallyParamsLawThreshold                                 = "tally_params_law_threshold"
	DepositParamsMinDepositFloor                            = "deposit_params_min_deposit_floor"
	DepositParamsMinDepositUpdatePeriod                     = "deposit_params_min_deposit_update_period"
	DepositParamsMinDepositSensitivityTargetDistance        = "deposit_params_min_deposit_sensitivity_target_distance"
	DepositParamsMinDepositIncreaseRatio                    = "deposit_params_min_deposit_increase_ratio"
	DepositParamsMinDepositDecreaseRatio                    = "deposit_params_min_deposit_decrease_ratio"
	DepositParamsTargetActiveProposals                      = "deposit_params_target_active_proposals"
	DepositParamsMinInitialDepositFloor                     = "deposit_params_min_initial_deposit_floor"
	DepositParamsMinInitialDepositUpdatePeriod              = "deposit_params_min_initial_deposit_update_period"
	DepositParamsMinInitialDepositSensitivityTargetDistance = "deposit_params_min_initial_deposit_sensitivity_target_distance"
	DepositParamsMinInitialDepositIncreaseRatio             = "deposit_params_min_initial_deposit_increase_ratio"
	DepositParamsMinInitialDepositDecreaseRatio             = "deposit_params_min_initial_deposit_decrease_ratio"
	DepositParamsMinInitialDepositTargetProposals           = "deposit_params_min_initial_deposit_target_proposals"
	MinDepositRatio                                         = "min_deposit_ratio"
	QuorumTimeout                                           = "quorum_timeout"
	MaxVotingPeriodExtension                                = "max_voting_period_extension"
	QuorumCheckCount                                        = "quorum_check_count"
	BurnDepositNoThreshold                                  = "burn_deposit_no_threshold"
	MinQuorum                                               = "min_quorum"
	MaxQuorum                                               = "max_quorum"
	MinConstitutionAmendmentQuorum                          = "min_constitution_amendment_quorum"
	MaxConstitutionAmendmentQuorum                          = "max_constitution_amendment_quorum"
	MinLawQuorum                                            = "min_law_quorum"
	MaxLawQuorum                                            = "max_law_quorum"
	ProposalCancelRate                                      = "proposal_cancel_rate"
	GovernorStatusChangePeriod                              = "governor_status_change_period"
	MinGovernorSelfDelegation                               = "min_governor_self_delegation"
)

// GenDepositParamsDepositPeriod returns randomized DepositParamsDepositPeriod
func GenDepositParamsDepositPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenDepositParamsMinDeposit returns randomized DepositParamsMinDeposit
func GenDepositParamsMinDeposit(r *rand.Rand) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simulation.RandIntBetween(r, 1, 1e3))))
}

// GenDepositMinInitialRatio returns randomized DepositMinInitialRatio
func GenDepositMinInitialDepositRatio(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(math.LegacyNewDec(100))
}

// GenVotingParamsVotingPeriod returns randomized VotingParamsVotingPeriod
func GenVotingParamsVotingPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenTallyParamsThreshold returns randomized TallyParamsThreshold
func GenTallyParamsThreshold(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 550, 700)), 3)
}

// GenMinDepositRatio returns randomized DepositMinRatio
func GenMinDepositRatio(*rand.Rand) math.LegacyDec {
	return math.LegacyMustNewDecFromStr("0.01")
}

// GenTallyParamsConstitutionalThreshold returns randomized TallyParamsConstitutionalThreshold
func GenTallyParamsConstitutionalThreshold(r *rand.Rand, minDec math.LegacyDec) math.LegacyDec {
	minimum := int(minDec.Mul(math.LegacyNewDec(1000)).RoundInt64())
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, minimum, 950)), 3)
}

// GenQuorumTimeout returns a randomized QuorumTimeout between 0 and votingPeriod
func GenQuorumTimeout(r *rand.Rand, votingPeriod time.Duration) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, int(votingPeriod.Seconds()))) * time.Second
}

// GenMaxVotingPeriodExtension returns a randomized MaxVotingPeriodExtension
// greater than votingPeriod-quorumTimout.
func GenMaxVotingPeriodExtension(r *rand.Rand, votingPeriod, quorumTimout time.Duration) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, int(votingPeriod.Seconds())))*time.Second + (votingPeriod - quorumTimout)
}

// GenQuorumCheckCount returns a randomized QuorumCheckCount between 0 and 30
func GenQuorumCheckCount(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 0, 30))
}

// GenMinQuorum returns a randomized MinQuorum between 0.15 and 0.25
func GenMinQuorum(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 15, 25)), 2)
}

// GenMaxQuorum returns a randomized MinQuorum between 0.75 and 0.85
func GenMaxQuorum(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 75, 85)), 2)
}

// GenDepositParamsMinDepositUpdatePeriod returns randomized DepositParamsMinDepositUpdatePeriod
func GenDepositParamsMinDepositUpdatePeriod(r *rand.Rand, votingPeriod time.Duration) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, int(votingPeriod.Seconds()))) * time.Second
}

// GenDepositParamsMinDepositSensitivityTargetDistance returns randomized DepositParamsMinDepositSensitivityTargetDistance
func GenDepositParamsMinDepositSensitivityTargetDistance(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 10))
}

// GenDepositParamsMinDepositChangeRatio returns randomized DepositParamsMinDepositChangeRatio
func GenDepositParamsMinDepositChangeRatio(r *rand.Rand, maximum, prec int) math.LegacyDec {
	if maximum <= 0 {
		return math.LegacyZeroDec()
	}
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 0, maximum)), int64(prec))
}

// GenDepositParamsTargetActiveProposals returns randomized DepositParamsTargetActiveProposals
func GenDepositParamsTargetActiveProposals(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 100))
}

// GenDepositParamsMinInitialDepositUpdatePeriod returns randomized DepositParamsMinInitialDepositUpdatePeriod
func GenDepositParamsMinInitialDepositUpdatePeriod(r *rand.Rand, depositPeriod time.Duration) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, int(depositPeriod.Seconds()))) * time.Second
}

// GenDepositParamsMinInitialDepositSensitivityTargetDistance returns randomized DepositParamsMinInitialDepositSensitivityTargetDistance
func GenDepositParamsMinInitialDepositSensitivityTargetDistance(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 10))
}

// GenDepositParamsMinInitialDepositChangeRatio returns randomized DepositParamsMinInitialDepositChangeRatio
func GenDepositParamsMinInitialDepositChangeRatio(r *rand.Rand, maximum, prec int) math.LegacyDec {
	if maximum <= 0 {
		return math.LegacyZeroDec()
	}
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 0, maximum)), int64(prec))
}

func GenDepositParamsMinInitialDepositTargetProposals(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 1, 100))
}

// GenBurnDepositNoThreshold returns a randomized BurnDepositNoThreshold between 0.5 and 0.95
func GenBurnDepositNoThreshold(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 500, 950)), 3)
}

// GenProposalCancelRate returns randomized ProposalCancelRate
func GenProposalCancelRate(r *rand.Rand) math.LegacyDec {
	return math.LegacyNewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(math.LegacyNewDec(100))
}

// GenMinGovernorSelfDelegation returns a randomized MinGovernorSelfDelegation
func GenMinGovernorSelfDelegation(r *rand.Rand) math.Int {
	return math.NewInt(int64(simulation.RandIntBetween(r, 1000, 10000000)))
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(simState *module.SimulationState) {
	startingProposalID := uint64(simState.Rand.Intn(100))
	startingParticipationEma := "0.25"

	// var minDeposit sdk.Coins
	// simState.AppParams.GetOrGenerate(
	//	 DepositParamsMinDeposit, &minDeposit, simState.Rand,
	//	func(r *rand.Rand) { minDeposit = GenDepositParamsMinDeposit(r) },
	//)

	var depositPeriod time.Duration
	simState.AppParams.GetOrGenerate(
		DepositParamsDepositPeriod, &depositPeriod, simState.Rand,
		func(r *rand.Rand) { depositPeriod = GenDepositParamsDepositPeriod(r) },
	)

	// var minInitialDepositRatio math.LegacyDec
	// simState.AppParams.GetOrGenerate(
	// 	 DepositMinInitialRatio, &minInitialDepositRatio, simState.Rand,
	// 	func(r *rand.Rand) { minInitialDepositRatio = GenDepositMinInitialDepositRatio(r) },
	// )

	var votingPeriod time.Duration
	simState.AppParams.GetOrGenerate(
		VotingParamsVotingPeriod, &votingPeriod, simState.Rand,
		func(r *rand.Rand) { votingPeriod = GenVotingParamsVotingPeriod(r) },
	)

	var threshold math.LegacyDec
	simState.AppParams.GetOrGenerate(
		TallyParamsThreshold, &threshold, simState.Rand,
		func(r *rand.Rand) { threshold = GenTallyParamsThreshold(r) },
	)

	var minDepositRatio math.LegacyDec
	simState.AppParams.GetOrGenerate(MinDepositRatio, &minDepositRatio, simState.Rand, func(r *rand.Rand) { minDepositRatio = GenMinDepositRatio(r) })

	var lawThreshold math.LegacyDec
	simState.AppParams.GetOrGenerate(
		TallyParamsLawThreshold, &lawThreshold, simState.Rand,
		func(r *rand.Rand) { lawThreshold = GenTallyParamsConstitutionalThreshold(r, threshold) },
	)

	var amendmentsThreshold math.LegacyDec
	simState.AppParams.GetOrGenerate(
		TallyParamsConstitutionAmendmentThreshold, &amendmentsThreshold, simState.Rand,
		func(r *rand.Rand) { amendmentsThreshold = GenTallyParamsConstitutionalThreshold(r, lawThreshold) },
	)

	var quorumTimout time.Duration
	simState.AppParams.GetOrGenerate(QuorumTimeout, &quorumTimout, simState.Rand, func(r *rand.Rand) { quorumTimout = GenQuorumTimeout(r, votingPeriod) })

	var maxVotingPeriodExtension time.Duration
	simState.AppParams.GetOrGenerate(MaxVotingPeriodExtension, &maxVotingPeriodExtension, simState.Rand, func(r *rand.Rand) {
		maxVotingPeriodExtension = GenMaxVotingPeriodExtension(r, votingPeriod, quorumTimout)
	})

	var quorumCheckCount uint64
	simState.AppParams.GetOrGenerate(QuorumCheckCount, &quorumCheckCount, simState.Rand, func(r *rand.Rand) { quorumCheckCount = GenQuorumCheckCount(r) })

	var minDepositFloor sdk.Coins
	simState.AppParams.GetOrGenerate(
		DepositParamsMinDepositFloor, &minDepositFloor, simState.Rand,
		func(r *rand.Rand) { minDepositFloor = GenDepositParamsMinDeposit(r) },
	)

	var minDepositUpdatePeriod time.Duration
	simState.AppParams.GetOrGenerate(
		DepositParamsMinDepositUpdatePeriod,
		&minDepositUpdatePeriod, simState.Rand,
		func(r *rand.Rand) { minDepositUpdatePeriod = GenDepositParamsMinDepositUpdatePeriod(r, votingPeriod) },
	)

	var minDepositSensitivityTargetDistance uint64
	simState.AppParams.GetOrGenerate(
		DepositParamsMinDepositSensitivityTargetDistance, &minDepositSensitivityTargetDistance, simState.Rand,
		func(r *rand.Rand) {
			minDepositSensitivityTargetDistance = GenDepositParamsMinDepositSensitivityTargetDistance(r)
		},
	)

	var minDepositIncreaseRatio math.LegacyDec
	simState.AppParams.GetOrGenerate(
		DepositParamsMinDepositIncreaseRatio, &minDepositIncreaseRatio, simState.Rand,
		func(r *rand.Rand) { minDepositIncreaseRatio = GenDepositParamsMinDepositChangeRatio(r, 300, 3) },
	)

	var minDepositDecreaseRatio math.LegacyDec
	simState.AppParams.GetOrGenerate(
		DepositParamsMinDepositDecreaseRatio, &minDepositDecreaseRatio, simState.Rand,
		func(r *rand.Rand) {
			minDepositDecreaseRatio = GenDepositParamsMinDepositChangeRatio(r,
				int(minDepositIncreaseRatio.MulInt64(1000).QuoInt64(2).TruncateInt64()), 3)
		},
	)

	var targetActiveProposals uint64
	simState.AppParams.GetOrGenerate(
		DepositParamsTargetActiveProposals, &targetActiveProposals, simState.Rand,
		func(r *rand.Rand) { targetActiveProposals = GenDepositParamsTargetActiveProposals(r) },
	)

	var minInitialDepositFloor sdk.Coins
	simState.AppParams.GetOrGenerate(
		DepositParamsMinInitialDepositFloor, &minInitialDepositFloor, simState.Rand,
		func(r *rand.Rand) {
			ratio := GenDepositMinInitialDepositRatio(r)
			minInitialDepositFloor = sdk.NewCoins()
			for _, coin := range minDepositFloor {
				minInitialDepositFloor = append(minInitialDepositFloor, sdk.NewCoin(coin.Denom, ratio.MulInt(coin.Amount).TruncateInt()))
			}
		},
	)

	var minInitialDepositUpdatePeriod time.Duration
	simState.AppParams.GetOrGenerate(
		DepositParamsMinInitialDepositUpdatePeriod, &minInitialDepositUpdatePeriod, simState.Rand,
		func(r *rand.Rand) {
			minInitialDepositUpdatePeriod = GenDepositParamsMinInitialDepositUpdatePeriod(r, depositPeriod)
		},
	)

	var minInitialDepositSensitivityTargetDistance uint64
	simState.AppParams.GetOrGenerate(
		DepositParamsMinInitialDepositSensitivityTargetDistance, &minInitialDepositSensitivityTargetDistance, simState.Rand,
		func(r *rand.Rand) {
			minInitialDepositSensitivityTargetDistance = GenDepositParamsMinInitialDepositSensitivityTargetDistance(r)
		},
	)

	var minInitialDepositIncreaseRatio math.LegacyDec
	simState.AppParams.GetOrGenerate(
		DepositParamsMinInitialDepositIncreaseRatio, &minInitialDepositIncreaseRatio, simState.Rand,
		func(r *rand.Rand) {
			minInitialDepositIncreaseRatio = GenDepositParamsMinInitialDepositChangeRatio(r, 300, 3)
		},
	)

	var minInitialDepositDecreaseRatio math.LegacyDec
	simState.AppParams.GetOrGenerate(
		DepositParamsMinInitialDepositDecreaseRatio, &minInitialDepositDecreaseRatio, simState.Rand,
		func(r *rand.Rand) {
			minInitialDepositDecreaseRatio = GenDepositParamsMinInitialDepositChangeRatio(r,
				int(minInitialDepositIncreaseRatio.MulInt64(1000).QuoInt64(2).TruncateInt64()), 3)
		},
	)

	var minInitialDepositTargetProposals uint64
	simState.AppParams.GetOrGenerate(
		DepositParamsMinInitialDepositTargetProposals, &minInitialDepositTargetProposals, simState.Rand,
		func(r *rand.Rand) {
			minInitialDepositTargetProposals = GenDepositParamsMinInitialDepositTargetProposals(r)
		},
	)

	var burnDepositNoThreshold math.LegacyDec
	simState.AppParams.GetOrGenerate(
		BurnDepositNoThreshold, &burnDepositNoThreshold, simState.Rand,
		func(r *rand.Rand) { burnDepositNoThreshold = GenBurnDepositNoThreshold(r) },
	)

	var minQuorum math.LegacyDec
	simState.AppParams.GetOrGenerate(MinQuorum, &minQuorum, simState.Rand, func(r *rand.Rand) { minQuorum = GenMinQuorum(r) })

	var maxQuorum math.LegacyDec
	simState.AppParams.GetOrGenerate(MaxQuorum, &maxQuorum, simState.Rand, func(r *rand.Rand) { maxQuorum = GenMaxQuorum(r) })

	var minConstitutionAmendmentQuorum math.LegacyDec
	simState.AppParams.GetOrGenerate(MinConstitutionAmendmentQuorum, &minConstitutionAmendmentQuorum, simState.Rand, func(r *rand.Rand) { minConstitutionAmendmentQuorum = GenMinQuorum(r) })

	var maxConstitutionAmendmentQuorum math.LegacyDec
	simState.AppParams.GetOrGenerate(MaxConstitutionAmendmentQuorum, &maxConstitutionAmendmentQuorum, simState.Rand, func(r *rand.Rand) { maxConstitutionAmendmentQuorum = GenMaxQuorum(r) })

	var minLawQuorum math.LegacyDec
	simState.AppParams.GetOrGenerate(MinLawQuorum, &minLawQuorum, simState.Rand, func(r *rand.Rand) { minLawQuorum = GenMinQuorum(r) })

	var maxLawQuorum math.LegacyDec
	simState.AppParams.GetOrGenerate(MaxLawQuorum, &maxLawQuorum, simState.Rand, func(r *rand.Rand) { maxLawQuorum = GenMaxQuorum(r) })

	var proposalCancelRate math.LegacyDec
	simState.AppParams.GetOrGenerate(ProposalCancelRate, &proposalCancelRate, simState.Rand, func(r *rand.Rand) { proposalCancelRate = GenProposalCancelRate(r) })

	var governorStatusChangePeriod time.Duration
	simState.AppParams.GetOrGenerate(
		GovernorStatusChangePeriod, &governorStatusChangePeriod, simState.Rand,
		func(r *rand.Rand) { governorStatusChangePeriod = GenDepositParamsDepositPeriod(r) },
	)

	var minGovernorSelfDelegation math.Int
	simState.AppParams.GetOrGenerate(
		MinGovernorSelfDelegation, &minGovernorSelfDelegation, simState.Rand,
		func(r *rand.Rand) { minGovernorSelfDelegation = GenMinGovernorSelfDelegation(r) },
	)

	govGenesis := v1.NewGenesisState(
		startingProposalID, startingParticipationEma, startingParticipationEma, startingParticipationEma,
		v1.NewParams(depositPeriod, votingPeriod, threshold.String(), amendmentsThreshold.String(), lawThreshold.String(),
			simState.Rand.Intn(2) == 0, simState.Rand.Intn(2) == 0, minDepositRatio.String(), quorumTimout,
			maxVotingPeriodExtension, quorumCheckCount, minDepositFloor, minDepositUpdatePeriod,
			minDepositSensitivityTargetDistance, minDepositIncreaseRatio.String(), minDepositDecreaseRatio.String(),
			targetActiveProposals, minInitialDepositFloor, minInitialDepositUpdatePeriod,
			minInitialDepositSensitivityTargetDistance, minInitialDepositIncreaseRatio.String(),
			minInitialDepositDecreaseRatio.String(), minInitialDepositTargetProposals,
			burnDepositNoThreshold.String(), maxQuorum.String(), minQuorum.String(),
			maxConstitutionAmendmentQuorum.String(), minConstitutionAmendmentQuorum.String(),
			maxLawQuorum.String(), minQuorum.String(),
			proposalCancelRate.String(), "",
			governorStatusChangePeriod, minGovernorSelfDelegation.String(),
		),
	)

	bz, err := json.MarshalIndent(&govGenesis, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(govGenesis)
}
