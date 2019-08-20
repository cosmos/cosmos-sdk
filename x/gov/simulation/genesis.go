package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// Simulation parameter constants
const (
	DepositParamsMinDeposit    = "deposit_params_min_deposit"
	DepositParamsDepositPeriod = "deposit_params_deposit_period"
	VotingParamsVotingPeriod   = "voting_params_voting_period"
	TallyParamsQuorum          = "tally_params_quorum"
	TallyParamsThreshold       = "tally_params_threshold"
	TallyParamsVeto            = "tally_params_veto"
)

// GenDepositParamsDepositPeriod randomized DepositParamsDepositPeriod
func GenDepositParamsDepositPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenDepositParamsMinDeposit randomized DepositParamsMinDeposit
func GenDepositParamsMinDeposit(r *rand.Rand) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simulation.RandIntBetween(r, 1, 1e3))))
}

// GenVotingParamsVotingPeriod randomized VotingParamsVotingPeriod
func GenVotingParamsVotingPeriod(r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenTallyParamsQuorum randomized TallyParamsQuorum
func GenTallyParamsQuorum(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 334, 500)), 3)
}

// GenTallyParamsThreshold randomized TallyParamsThreshold
func GenTallyParamsThreshold(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 450, 550)), 3)
}

// GenTallyParamsVeto randomized TallyParamsVeto
func GenTallyParamsVeto(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 250, 334)), 3)
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(input *module.GeneratorInput) {

	var (
		minDeposit    sdk.Coins
		depositPeriod time.Duration
		votingPeriod  time.Duration
		quorum        sdk.Dec
		threshold     sdk.Dec
		veto          sdk.Dec
	)

	startingProposalID := uint64(input.R.Intn(100))

	input.AppParams.GetOrGenerate(input.Cdc, DepositParamsMinDeposit, &minDeposit, input.R,
		func(r *rand.Rand) { minDeposit = GenDepositParamsMinDeposit(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, DepositParamsDepositPeriod, &depositPeriod, input.R,
		func(r *rand.Rand) { depositPeriod = GenDepositParamsDepositPeriod(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, VotingParamsVotingPeriod, &votingPeriod, input.R,
		func(r *rand.Rand) { votingPeriod = GenVotingParamsVotingPeriod(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, TallyParamsQuorum, &quorum, input.R,
		func(r *rand.Rand) { quorum = GenTallyParamsQuorum(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, TallyParamsThreshold, &threshold, input.R,
		func(r *rand.Rand) { threshold = GenTallyParamsThreshold(input.R) })

	input.AppParams.GetOrGenerate(input.Cdc, TallyParamsVeto, &veto, input.R,
		func(r *rand.Rand) { veto = GenTallyParamsVeto(input.R) })

	govGenesis := types.NewGenesisState(
		startingProposalID,
		types.NewDepositParams(minDeposit, depositPeriod),
		types.NewVotingParams(votingPeriod),
		types.NewTallyParams(quorum, threshold, veto),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, govGenesis))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(govGenesis)
}
