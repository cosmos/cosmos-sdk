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
	DepositParamsMinDeposit  = "deposit_params_min_deposit"
	VotingParamsVotingPeriod = "voting_params_voting_period"
	TallyParamsQuorum        = "tally_params_quorum"
	TallyParamsThreshold     = "tally_params_threshold"
	TallyParamsVeto          = "tally_params_veto"
)

// GenGovGenesisState generates a random GenesisState for gov
func GenGovGenesisState(cdc *codec.Codec, r *rand.Rand, ap simulation.AppParams, genesisState map[string]json.RawMessage) {
	startingProposalID := uint64(r.Intn(100))
	
	// period
	// TODO: create 2 separate periods for deposit and voting
	var period time.Duration
	ap.GetOrGenerate(cdc, VotingParamsVotingPeriod, &period, r,
		func(r *rand.Rand) {
			period = time.Duration(RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
		})

	// deposit params
	var minDeposit sdk.Coins
		ap.GetOrGenerate(cdc, DepositParamsMinDeposit, &minDeposit, r,
			func(r *rand.Rand) {
				minDeposit = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(RandIntBetween(r, 1, 1e3))))
			})

	// tally params
	var quorum sdk.Dec
	ap.GetOrGenerate(cdc, TallyParamsQuorum, &quorum, r,
		func(r *rand.Rand) {
			quorum = sdk.NewDecWithPrec(int64(RandIntBetween(r, 334, 500)), 3)
		})

	var threshold sdk.Dec
	ap.GetOrGenerate(cdc, TallyParamsThreshold, &threshold, r,
		func(r *rand.Rand) {
			threshold = sdk.NewDecWithPrec(int64(RandIntBetween(r, 450, 550)), 3)
		})

	var veto sdk.Dec
	ap.GetOrGenerate(cdc, TallyParamsVeto, &veto, r,
		func(r *rand.Rand) {
			veto = sdk.NewDecWithPrec(int64(RandIntBetween(r, 250, 334)), 3)
		})

	govGenesis := gov.NewGenesisState(
		startingProposalID,
		gov.NewDepositParams(minDeposit, period),
		gov.NewVotingParams(period),
		gov.NewTallyParams(quorum, threshold, veto),
	)

	fmt.Printf("Selected randomly generated governance parameters:\n%s\n", codec.MustMarshalJSONIndent(cdc, govGenesis))
	genesisState[gov.ModuleName] = cdc.MustMarshalJSON(govGenesis)
}
