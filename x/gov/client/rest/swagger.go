package rest

import (
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov/client/utils"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// Concrete Swagger types used to generate REST documentation. Note, these types
// are not actually used but since all queries return a generic JSON raw message,
// they enabled typed documentation.
//
// nolint: deadcode unused
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
		Height int64          `json:"height"`
		Result utils.Proposer `json:"result"`
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

	postProposal struct {
		Msgs       []types.MsgSubmitProposal `json:"msg" yaml:"msg"`
		Fee        auth.StdFee               `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature       `json:"signatures" yaml:"signatures"`
		Memo       string                    `json:"memo" yaml:"memo"`
	}

	postDeposit struct {
		Msgs       []types.MsgDeposit  `json:"msg" yaml:"msg"`
		Fee        auth.StdFee         `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
		Memo       string              `json:"memo" yaml:"memo"`
	}

	postVote struct {
		Msgs       []types.MsgVote     `json:"msg" yaml:"msg"`
		Fee        auth.StdFee         `json:"fee" yaml:"fee"`
		Signatures []auth.StdSignature `json:"signatures" yaml:"signatures"`
		Memo       string              `json:"memo" yaml:"memo"`
	}
)
