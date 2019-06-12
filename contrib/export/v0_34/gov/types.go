package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"time"
)

type GenesisState struct {
	StartingProposalID uint64                `json:"starting_proposal_id"`
	Deposits           []DepositWithMetadata `json:"deposits"`
	Votes              []VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal            `json:"proposals"`
	DepositParams      gov.DepositParams     `json:"deposit_params"`
	VotingParams       gov.VotingParams      `json:"voting_params"`
	TallyParams        gov.TallyParams       `json:"tally_params"`
}

type DepositWithMetadata struct {
	ProposalID uint64      `json:"proposal_id"`
	Deposit    gov.Deposit `json:"deposit"`
}

type VoteWithMetadata struct {
	ProposalID uint64   `json:"proposal_id"`
	Vote       gov.Vote `json:"vote"`
}

type Proposal struct {
	gov.Content `json:"content"` // Proposal content interface

	ProposalID       uint64             `json:"proposal_id"`        //  ID of the proposal
	Status           gov.ProposalStatus `json:"proposal_status"`    // Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult gov.TallyResult    `json:"final_tally_result"` // Result of Tallys

	SubmitTime     time.Time `json:"submit_time"`      // Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins `json:"total_deposit"`    // Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` // Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

type Proposals []Proposal
