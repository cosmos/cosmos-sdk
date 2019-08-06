package simulation

// DONTCOVER

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
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
func GenDepositParamsDepositPeriod(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (period time.Duration) {
	ap.GetOrGenerate(cdc, DepositParamsDepositPeriod, &period, r,
		func(r *rand.Rand) {
			period = time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
		})
	return
}

// GenDepositParamsMinDeposit randomized DepositParamsMinDeposit
func GenDepositParamsMinDeposit(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (minDeposit sdk.Coins) {
	ap.GetOrGenerate(cdc, DepositParamsMinDeposit, &minDeposit, r,
		func(r *rand.Rand) {
			minDeposit = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(simulation.RandIntBetween(r, 1, 1e3))))
		})
	return
}

// GenVotingParamsVotingPeriod randomized VotingParamsVotingPeriod
func GenVotingParamsVotingPeriod(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (period time.Duration) {
	ap.GetOrGenerate(cdc, VotingParamsVotingPeriod, &period, r,
		func(r *rand.Rand) {
			period = time.Duration(simulation.RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
		})
	return
}

// GenTallyParamsQuorum randomized TallyParamsQuorum
func GenTallyParamsQuorum(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (quorum sdk.Dec) {
	ap.GetOrGenerate(cdc, TallyParamsQuorum, &quorum, r,
		func(r *rand.Rand) {
			quorum = sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 334, 500)), 3)
		})
	return
}

// GenTallyParamsThreshold randomized TallyParamsThreshold
func GenTallyParamsThreshold(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (threshold sdk.Dec) {
	ap.GetOrGenerate(cdc, TallyParamsThreshold, &threshold, r,
		func(r *rand.Rand) {
			threshold = sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 450, 550)), 3)
		})
	return
}

// GenTallyParamsVeto randomized TallyParamsVeto
func GenTallyParamsVeto(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams) (veto sdk.Dec) {
	ap.GetOrGenerate(cdc, TallyParamsVeto, &veto, r,
		func(r *rand.Rand) {
			veto = sdk.NewDecWithPrec(int64(simulation.RandIntBetween(r, 250, 334)), 3)
		})
	return
}

// GenGovGenesisState generates a random GenesisState for gov
func GenGovGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	startingProposalID := uint64(r.Intn(100))

	minDeposit := GenDepositParamsMinDeposit(cdc, r, ap)
	depositPeriod := GenDepositParamsDepositPeriod(cdc, r, ap)
	votingPeriod := GenVotingParamsVotingPeriod(cdc, r, ap)
	quorum := GenTallyParamsQuorum(cdc, r, ap)
	threshold := GenTallyParamsThreshold(cdc, r, ap)
	veto := GenTallyParamsVeto(cdc, r, ap)

	govGenesis := gov.NewGenesisState(
		startingProposalID,
		gov.NewDepositParams(minDeposit, depositPeriod),
		gov.NewVotingParams(votingPeriod),
		gov.NewTallyParams(quorum, threshold, veto),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, govGenesis))
	genesisState[gov.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}
