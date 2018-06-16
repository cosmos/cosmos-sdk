package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Type that represents Status as a byte
type VoteStatus = byte

// Type that represents Proposal Type as a byte
type ProposalKind = byte

//nolint
const (
	StatusDepositPeriod VoteStatus = 0x01
	StatusVotingPeriod  VoteStatus = 0x02
	StatusPassed        VoteStatus = 0x03
	StatusRejected      VoteStatus = 0x04

	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

//-----------------------------------------------------------
// Proposal interface
type Proposal interface {
	GetProposalID() int64
	SetProposalID(int64)

	GetTitle() string
	SetTitle(string)

	GetDescription() string
	SetDescription(string)

	GetProposalType() ProposalKind
	SetProposalType(ProposalKind)

	GetStatus() VoteStatus
	SetStatus(VoteStatus)

	GetSubmitBlock() int64
	SetSubmitBlock(int64)

	GetTotalDeposit() sdk.Coins
	SetTotalDeposit(sdk.Coins)

	GetVotingStartBlock() int64
	SetVotingStartBlock(int64)
}

// checks if two proposals are equal
func ProposalEqual(proposalA Proposal, proposalB Proposal) bool {
	if proposalA.GetProposalID() != proposalB.GetProposalID() ||
		proposalA.GetTitle() != proposalB.GetTitle() ||
		proposalA.GetDescription() != proposalB.GetDescription() ||
		proposalA.GetProposalType() != proposalB.GetProposalType() ||
		proposalA.GetStatus() != proposalB.GetStatus() ||
		proposalA.GetSubmitBlock() != proposalB.GetSubmitBlock() ||
		!(proposalA.GetTotalDeposit().IsEqual(proposalB.GetTotalDeposit())) ||
		proposalA.GetVotingStartBlock() != proposalB.GetVotingStartBlock() {
		return false
	}
	return true
}

//-----------------------------------------------------------
// Text Proposals
type TextProposal struct {
	ProposalID   int64        `json:"proposal_id"`   //  ID of the proposal
	Title        string       `json:"title"`         //  Title of the proposal
	Description  string       `json:"description"`   //  Description of the proposal
	ProposalType ProposalKind `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}

	Status VoteStatus `json:"string"` //  Status of the Proposal {Pending, Active, Passed, Rejected}

	SubmitBlock  int64     `json:"submit_block"`  //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit sdk.Coins `json:"total_deposit"` //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartBlock int64 `json:"voting_start_block"` //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

// Implements Proposal Interface
var _ Proposal = (*TextProposal)(nil)

// nolint
func (tp TextProposal) GetProposalID() int64                       { return tp.ProposalID }
func (tp *TextProposal) SetProposalID(proposalID int64)            { tp.ProposalID = proposalID }
func (tp TextProposal) GetTitle() string                           { return tp.Title }
func (tp *TextProposal) SetTitle(title string)                     { tp.Title = title }
func (tp TextProposal) GetDescription() string                     { return tp.Description }
func (tp *TextProposal) SetDescription(description string)         { tp.Description = description }
func (tp TextProposal) GetProposalType() ProposalKind              { return tp.ProposalType }
func (tp *TextProposal) SetProposalType(proposalType ProposalKind) { tp.ProposalType = proposalType }
func (tp TextProposal) GetStatus() VoteStatus                      { return tp.Status }
func (tp *TextProposal) SetStatus(status VoteStatus)               { tp.Status = status }
func (tp TextProposal) GetSubmitBlock() int64                      { return tp.SubmitBlock }
func (tp *TextProposal) SetSubmitBlock(submitBlock int64)          { tp.SubmitBlock = submitBlock }
func (tp TextProposal) GetTotalDeposit() sdk.Coins                 { return tp.TotalDeposit }
func (tp *TextProposal) SetTotalDeposit(totalDeposit sdk.Coins)    { tp.TotalDeposit = totalDeposit }
func (tp TextProposal) GetVotingStartBlock() int64                 { return tp.VotingStartBlock }
func (tp *TextProposal) SetVotingStartBlock(votingStartBlock int64) {
	tp.VotingStartBlock = votingStartBlock
}

// Current Active Proposals
type ProposalQueue []int64

// ProposalTypeToString for pretty prints of ProposalType
func ProposalTypeToString(proposalType ProposalKind) string {
	switch proposalType {
	case 0x00:
		return "Text"
	case 0x01:
		return "ParameterChange"
	case 0x02:
		return "SoftwareUpgrade"
	default:
		return ""
	}
}

func validProposalType(proposalType ProposalKind) bool {
	if proposalType == ProposalTypeText ||
		proposalType == ProposalTypeParameterChange ||
		proposalType == ProposalTypeSoftwareUpgrade {
		return true
	}
	return false
}

// String to proposalType byte.  Returns ff if invalid.
func StringToProposalType(str string) (ProposalKind, sdk.Error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
	case "ParameterChange":
		return ProposalTypeParameterChange, nil
	case "SoftwareUpgrade":
		return ProposalTypeSoftwareUpgrade, nil
	default:
		return ProposalKind(0xff), ErrInvalidProposalType(DefaultCodespace, str)
	}
}

// StatusToString for pretty prints of Status
func StatusToString(status VoteStatus) string {
	switch status {
	case StatusDepositPeriod:
		return "DepositPeriod"
	case StatusVotingPeriod:
		return "VotingPeriod"
	case StatusPassed:
		return "Passed"
	case StatusRejected:
		return "Rejected"
	default:
		return ""
	}
}

// StatusToString for pretty prints of Status
func StringToStatus(status string) VoteStatus {
	switch status {
	case "DepositPeriod":
		return StatusDepositPeriod
	case "VotingPeriod":
		return StatusVotingPeriod
	case "Passed":
		return StatusPassed
	case "Rejected":
		return StatusRejected
	default:
		return VoteStatus(0xff)
	}
}

//-----------------------------------------------------------
// Rest Proposals
type ProposalRest struct {
	ProposalID       int64     `json:"proposal_id"`        //  ID of the proposal
	Title            string    `json:"title"`              //  Title of the proposal
	Description      string    `json:"description"`        //  Description of the proposal
	ProposalType     string    `json:"proposal_type"`      //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
	Status           string    `json:"string"`             //  Status of the Proposal {Pending, Active, Passed, Rejected}
	SubmitBlock      int64     `json:"submit_block"`       //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit     sdk.Coins `json:"total_deposit"`      //  Current deposit on this proposal. Initial value is set at InitialDeposit
	VotingStartBlock int64     `json:"voting_start_block"` //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

// Turn any Proposal to a ProposalRest
func ProposalToRest(proposal Proposal) ProposalRest {
	return ProposalRest{
		ProposalID:       proposal.GetProposalID(),
		Title:            proposal.GetTitle(),
		Description:      proposal.GetDescription(),
		ProposalType:     ProposalTypeToString(proposal.GetProposalType()),
		Status:           StatusToString(proposal.GetStatus()),
		SubmitBlock:      proposal.GetSubmitBlock(),
		TotalDeposit:     proposal.GetTotalDeposit(),
		VotingStartBlock: proposal.GetVotingStartBlock(),
	}
}
