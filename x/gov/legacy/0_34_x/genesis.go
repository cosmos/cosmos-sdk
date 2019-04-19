package legacy

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// GenesisState defines legacy governance module genesis state for version
// 0.34.x.
type GenesisState struct {
	StartingProposalID uint64                    `json:"starting_proposal_id"`
	Deposits           []gov.DepositWithMetadata `json:"deposits"`
	Votes              []gov.VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal                `json:"proposals"`
	DepositParams      gov.DepositParams         `json:"deposit_params"`
	VotingParams       gov.VotingParams          `json:"voting_params"`
	TallyParams        gov.TallyParams           `json:"tally_params"`
}

// InitGenesis initializes genesis state in a (non-legacy) governance keeper
// with legacy GenesisState.
func InitGenesis(ctx sdk.Context, k gov.Keeper, data GenesisState) {
	gov.InitGenesis(ctx, k, MigrateGenesis(data))
}

// MigrateGenesis transforms data from legacy GenesisState to non-legacy
// GenesisState.
func MigrateGenesis(data GenesisState) (res gov.GenesisState) {
	res = gov.GenesisState{
		StartingProposalID: data.StartingProposalID,
		Deposits:           data.Deposits,
		Votes:              data.Votes,
		Proposals:          make([]gov.Proposal, len(data.Proposals)),
		DepositParams:      data.DepositParams,
		VotingParams:       data.VotingParams,
		TallyParams:        data.TallyParams,
	}

	var err error
	for i, p := range data.Proposals {
		res.Proposals[i], err = p.Migrate()
		if err != nil {
			panic(err)
		}
	}

	return
}
