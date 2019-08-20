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
	DepositParamsDepositPeriod = "deposit_params_deposit_period"
	DepositParamsMinDeposit    = "deposit_params_min_deposit"
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
	startingProposalID := uint64(input.R.Intn(100))

	minDeposit := GenDepositParamsMinDeposit(input.R)
	depositPeriod := GenDepositParamsDepositPeriod(input.R)
	votingPeriod := GenVotingParamsVotingPeriod(input.R)
	quorum := GenTallyParamsQuorum(input.R)
	threshold := GenTallyParamsThreshold(input.R)
	veto := GenTallyParamsVeto(input.R)

	govGenesis := types.NewGenesisState(
		startingProposalID,
		types.NewDepositParams(minDeposit, depositPeriod),
		types.NewVotingParams(votingPeriod),
		types.NewTallyParams(quorum, threshold, veto),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(input.Cdc, govGenesis))
	input.GenState[types.ModuleName] = input.Cdc.MustMarshalJSON(govGenesis)
}
