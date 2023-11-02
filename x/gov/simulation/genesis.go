package simulation

import (
	"encoding/json"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/x/gov/types"
	v1 "cosmossdk.io/x/gov/types/v1"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
)

// Simulation parameter constants
const (
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
func GenDepositMinInitialDepositRatio(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(sdkmath.LegacyNewDec(100))
}

// GenProposalCancelRate returns randomized ProposalCancelRate
func GenProposalCancelRate(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDec(int64(simulation.RandIntBetween(r, 0, 99))).Quo(sdkmath.LegacyNewDec(100))
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
func GenExpeditedThreshold(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, tallyNonExpeditedMax, 550)), 3)
}

// GenVeto returns randomized Veto
func GenVeto(r *rand.Rand) sdkmath.LegacyDec {
	return sdkmath.LegacyNewDecWithPrec(int64(simulation.RandIntBetween(r, 250, 334)), 3)
}

// RandomizedGenState generates a random GenesisState for gov.
func RandomizedGenState(r *rand.Rand, genState map[string]json.RawMessage, cdc codec.JSONCodec, bondDenom string) {
	startingProposalID := uint64(r.Intn(100))

	minDeposit := GenMinDeposit(r, bondDenom)
	expeditedMinDeposit := GenExpeditedMinDeposit(r, bondDenom)
	depositPeriod := GenDepositPeriod(r)
	minInitialDepositRatio := GenDepositMinInitialDepositRatio(r)
	proposalCancelRate := GenProposalCancelRate(r)
	votingPeriod := GenVotingPeriod(r)
	expeditedVotingPeriod := GenExpeditedVotingPeriod(r)
	quorum := GenQuorum(r)
	threshold := GenThreshold(r)
	expitedVotingThreshold := GenExpeditedThreshold(r)
	veto := GenVeto(r)

	minDepositRatio := sdkmath.LegacyMustNewDecFromStr("0.01")

	govGenesis := v1.NewGenesisState(
		startingProposalID,
		v1.NewParams(minDeposit, expeditedMinDeposit, depositPeriod, votingPeriod, expeditedVotingPeriod, quorum.String(), threshold.String(), expitedVotingThreshold.String(), veto.String(), minInitialDepositRatio.String(), proposalCancelRate.String(), "", r.Intn(2) == 0, r.Intn(2) == 0, r.Intn(2) == 0, minDepositRatio.String()),
	)

	genState[types.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}
