package simulation

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Simulation parameter constants
const (
	DepositParamsMinDeposit  = "deposit_params_min_deposit"
	VotingParamsVotingPeriod = "voting_params_voting_period"
	TallyParamsQuorum        = "tally_params_quorum"
	TallyParamsThreshold     = "tally_params_threshold"
	TallyParamsVeto          = "tally_params_veto"
)

// GenParams generates random gov parameters
func GenParams(paramSims map[string]func(r *rand.Rand) interface{}) {
	paramSims[DepositParamsMinDeposit] = func(r *rand.Rand) interface{} {
		return sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(RandIntBetween(r, 1, 1e3)))}
	}

	paramSims[VotingParamsVotingPeriod] = func(r *rand.Rand) interface{} {
		return time.Duration(RandIntBetween(r, 1, 2*60*60*24*2)) * time.Second
	}

	paramSims[TallyParamsQuorum] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(int64(RandIntBetween(r, 334, 500)), 3)
	}

	paramSims[TallyParamsThreshold] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(int64(RandIntBetween(r, 450, 550)), 3)
	}

	paramSims[TallyParamsVeto] = func(r *rand.Rand) interface{} {
		return sdk.NewDecWithPrec(int64(RandIntBetween(r, 250, 334)), 3)
	}
}
