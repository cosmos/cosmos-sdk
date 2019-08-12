package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
func GenDepositParamsDepositPeriod(cdc *codec.Codec, r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenDepositParamsMinDeposit randomized DepositParamsMinDeposit
func GenDepositParamsMinDeposit(cdc *codec.Codec, r *rand.Rand) sdk.Coins {
	return sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simulation.RandIntBetween(r, 1, 1e3))))
}

// GenVotingParamsVotingPeriod randomized VotingParamsVotingPeriod
func GenVotingParamsVotingPeriod(cdc *codec.Codec, r *rand.Rand) time.Duration {
	return time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
}

// GenTallyParamsQuorum randomized TallyParamsQuorum
func GenTallyParamsQuorum(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 334, 500)), 3)
}

// GenTallyParamsThreshold randomized TallyParamsThreshold
func GenTallyParamsThreshold(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 450, 550)), 3)
}

// GenTallyParamsVeto randomized TallyParamsVeto
func GenTallyParamsVeto(cdc *codec.Codec, r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 250, 334)), 3)
}

// RandomizedGenState generates a random GenesisState for gov
func RandomizedGenState(cdc *codec.Codec, r *rand.Rand, genesisState map[string]json.RawMessage) {
	startingProposalID := uint64(r.Intn(100))

	minDeposit := GenDepositParamsMinDeposit(cdc, r)
	depositPeriod := GenDepositParamsDepositPeriod(cdc, r)
	votingPeriod := GenVotingParamsVotingPeriod(cdc, r)
	quorum := GenTallyParamsQuorum(cdc, r)
	threshold := GenTallyParamsThreshold(cdc, r)
	veto := GenTallyParamsVeto(cdc, r)

	govGenesis := types.NewGenesisState(
		startingProposalID,
		types.NewDepositParams(minDeposit, depositPeriod),
		types.NewVotingParams(votingPeriod),
		types.NewTallyParams(quorum, threshold, veto),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, govGenesis))
	genesisState[types.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}
