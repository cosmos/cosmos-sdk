package rest

import (
	gcutils "github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
// nolint: deadcode
type (
	queryDeposits struct {
		Height int64           `json:"height"`
		Result []types.Deposit `json:"result"`
	}

	queryProposal struct {
		Height int64          `json:"height"`
		Result types.Proposal `json:"result"`
	}

	queryProposer struct {
		Height int64            `json:"height"`
		Result gcutils.Proposer `json:"result"`
	}

	queryDeposit struct {
		Height int64         `json:"height"`
		Result types.Deposit `json:"result"`
	}

	queryVote struct {
		Height int64      `json:"height"`
		Result types.Vote `json:"result"`
	}

	queryVotesOnProposal struct {
		Height int64       `json:"height"`
		Result types.Votes `json:"result"`
	}

	queryProposals struct {
		Height int64            `json:"height"`
		Result []types.Proposal `json:"result"`
	}

	queryTally struct {
		Height int64             `json:"height"`
		Result types.TallyResult `json:"result"`
	}
)
