package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

//nolint
const (
	StatusDepositPeriod = "DepositPeriod"
	StatusVotingPeriod  = "VotingPeriod"
	StatusPassed        = "Passed"
	StatusRejected      = "Rejected"
)

var (
	proposalTypes = []string{"Text", "ParameterChange", "SoftwareUpgrade"}
)

// Vote
type Vote struct {
	Voter      sdk.Address `json:"voter"`       //  address of the voter
	ProposalID int64       `json:"proposal_id"` //  proposalID of the proposal
	Option     string      `json:"option"`      //  option from OptionSet chosen by the voter
}

//-----------------------------------------------------------
// Proposal Interface
type Proposal interface {
	GetProposalID() int64
	SetProposalID(int64)

	GetTitle() string
	SetTitle(string)

	GetDescription() string
	SetDescription(string)

	GetProposalType() string
	SetProposalType(string)

	GetStatus() string
	SetStatus(string)

	GetSubmitBlock() int64
	SetSubmitBlock(int64)

	GetTotalDeposit() sdk.Coins
	SetTotalDeposit(sdk.Coins)

	GetVotingStartBlock() int64
	SetVotingStartBlock(int64)
}

//-----------------------------------------------------------
// Text Proposal
type TextProposal struct {
	ProposalID   int64  `json:"proposal_id"`   //  ID of the proposal
	Title        string `json:"title"`         //  Title of the proposal
	Description  string `json:"description"`   //  Description of the proposal
	ProposalType string `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}

	Status string `json:"string"` //  Status of the Proposal {Pending, Active, Passed, Rejected}

	SubmitBlock  int64     `json:"submit_block"`  //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit sdk.Coins `json:"total_deposit"` //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartBlock int64 `json:"voting_start_block"` //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

// Implements Proposal Interface
var _ Proposal = (*TextProposal)(nil)

// nolint
func (tp TextProposal) GetProposalID() int64                    { return tp.ProposalID }
func (tp *TextProposal) SetProposalID(proposalID int64)         { tp.ProposalID = proposalID }
func (tp TextProposal) GetTitle() string                        { return tp.Title }
func (tp *TextProposal) SetTitle(title string)                  { tp.Title = title }
func (tp TextProposal) GetDescription() string                  { return tp.Description }
func (tp *TextProposal) SetDescription(description string)      { tp.Description = description }
func (tp TextProposal) GetProposalType() string                 { return tp.ProposalType }
func (tp *TextProposal) SetProposalType(proposalType string)    { tp.ProposalType = proposalType }
func (tp TextProposal) GetStatus() string                       { return tp.Status }
func (tp *TextProposal) SetStatus(status string)                { tp.Status = status }
func (tp TextProposal) GetSubmitBlock() int64                   { return tp.SubmitBlock }
func (tp *TextProposal) SetSubmitBlock(submitBlock int64)       { tp.SubmitBlock = submitBlock }
func (tp TextProposal) GetTotalDeposit() sdk.Coins              { return tp.TotalDeposit }
func (tp *TextProposal) SetTotalDeposit(totalDeposit sdk.Coins) { tp.TotalDeposit = totalDeposit }
func (tp TextProposal) GetVotingStartBlock() int64              { return tp.VotingStartBlock }
func (tp *TextProposal) SetVotingStartBlock(votingStartBlock int64) {
	tp.VotingStartBlock = votingStartBlock
}

//-----------------------------------------------------------
// Procedures

// Procedure around Deposits for governance
type DepositProcedure struct {
	MinDeposit       sdk.Coins `json:"min_deposit"`        //  Minimum deposit for a proposal to enter voting period.
	MaxDepositPeriod int64     `json:"max_deposit_period"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}

// Procedure around Tallying votes in governance
type TallyingProcedure struct {
	Threshold         sdk.Rat `json:"threshold"`          //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
	Veto              sdk.Rat `json:"veto"`               //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
	GovernancePenalty sdk.Rat `json:"governance_penalty"` //  Penalty if validator does not vote
}

// Procedure around Voting in governance
type VotingProcedure struct {
	VotingPeriod int64 `json:"voting_period"` //  Length of the voting period.
}

// List of valid proposal types
type ProposalTypes []string

func validProposalType(proposalType string) bool {
	for _, p := range proposalTypes {
		if p == proposalType {
			return true
		}
	}
	return false
}

// Deposit
type Deposit struct {
	Depositer sdk.Address `json:"depositer"` //  Address of the depositer
	Amount    sdk.Coins   `json:"amount"`    //  Deposit amount
}

// validatorGovInfo used for tallying
type validatorGovInfo struct {
	ValidatorInfo stake.Validator //  Voting power of validator when proposal enters voting period
	Minus         sdk.Rat         //  Minus of validator, used to compute validator's voting power
	Vote          string          // Vote of the validator
	Power         sdk.Rat         // Power of a Validator
}

// Current Active Proposals
type ProposalQueue []int64
