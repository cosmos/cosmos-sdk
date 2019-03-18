package gov

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: this is ugly...
// we should make the legacy proposals automatically migrated into the new format
// so we can remove this file in the next chain upgrade
// one of the way is when GetProposal() detects a LegacyProposal
// it converts into & sets a new Proposal
// but GetProposal() happens only in AddDeposit(), AddVote()
// (which fails to be executed after deposit/voting period)
// and queriers (which does not mutate the state)

var legacyCdc = codec.New()

func init() {
	legacyCdc.RegisterConcrete(legacyProposal{}, "gov/TextProposal", nil)
}

type ProposalKind byte

const (
	legacyProposalTypeText            ProposalKind = 0x01
	legacyProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

type legacyProposal struct {
	ProposalID   uint64       `json:"proposal_id"`   //  ID of the proposal
	Title        string       `json:"title"`         //  Title of the proposal
	Description  string       `json:"description"`   //  Description of the proposal
	ProposalType ProposalKind `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}

	Status           ProposalStatus `json:"proposal_status"`    //  Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult TallyResult    `json:"final_tally_result"` //  Result of Tallys

	SubmitTime     time.Time `json:"submit_time"`      //  Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins `json:"total_deposit"`    //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` //  Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

func proposalFromLegacy(p legacyProposal) (res Proposal) {
	res = Proposal{
		ProposalID:       p.ProposalID,
		Status:           p.Status,
		FinalTallyResult: p.FinalTallyResult,
		SubmitTime:       p.SubmitTime,
		DepositEndTime:   p.DepositEndTime,
		TotalDeposit:     p.TotalDeposit,
		VotingStartTime:  p.VotingStartTime,
		VotingEndTime:    p.VotingEndTime,
	}

	switch p.ProposalType {
	case legacyProposalTypeText:
		res.Content = NewTextProposal(p.Title, p.Description)
	case legacyProposalTypeSoftwareUpgrade:
		res.Content = NewSoftwareUpgradeProposal(p.Title, p.Description)
	}

	return
}
