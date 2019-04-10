package gov

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

// Proposal is a struct used by gov module internally
// embedds ProposalContent with additional fields to record the status of the proposal process
type Proposal struct {
	proposal.Content `json:"content"` // Proposal content interface

	ProposalID uint64 `json:"id"` //  ID of the proposal

	Status           ProposalStatus `json:"proposal_status"`    //  Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult TallyResult    `json:"final_tally_result"` //  Result of Tallys

	SubmitTime     time.Time `json:"submit_time"`      //  Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins `json:"total_deposit"`    //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` //  Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

func NewProposal(content proposal.Content, id uint64, submitTime, depositEndTime time.Time) Proposal {
	return Proposal{
		Content:          content,
		ProposalID:       id,
		Status:           StatusDepositPeriod,
		FinalTallyResult: EmptyTallyResult(),
		TotalDeposit:     sdk.NewCoins(),
		SubmitTime:       submitTime,
		DepositEndTime:   depositEndTime,
	}
}

// nolint
func (p Proposal) String() string {
	return fmt.Sprintf(`Proposal %d:
  Title:              %s
  Type:               %s
  Status:             %s
  Submit Time:        %s
  Deposit End Time:   %s
  Total Deposit:      %s
  Voting Start Time:  %s
  Voting End Time:    %s
  Description:        %s`,
		p.ProposalID, p.GetTitle(), p.ProposalType(),
		p.Status, p.SubmitTime, p.DepositEndTime,
		p.TotalDeposit, p.VotingStartTime, p.VotingEndTime, p.GetDescription(),
	)
}

// Proposals is an array of proposal
type Proposals []Proposal

// nolint
func (p Proposals) String() string {
	out := "ID - (Status) [Type] Title\n"
	for _, prop := range p {
		out += fmt.Sprintf("%d - (%s) [%s] %s\n",
			prop.ProposalID, prop.Status,
			prop.ProposalType(), prop.GetTitle())
	}
	return strings.TrimSpace(out)
}

// ProposalQueue
type ProposalQueue []uint64

// ProposalStatus

// Type that represents Proposal Status as a byte
type ProposalStatus byte

//nolint
const (
	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
)

// ProposalStatusToString turns a string into a ProposalStatus
func ProposalStatusFromString(str string) (ProposalStatus, error) {
	switch str {
	case "DepositPeriod":
		return StatusDepositPeriod, nil
	case "VotingPeriod":
		return StatusVotingPeriod, nil
	case "Passed":
		return StatusPassed, nil
	case "Rejected":
		return StatusRejected, nil
	case "":
		return StatusNil, nil
	default:
		return ProposalStatus(0xff), fmt.Errorf("'%s' is not a valid proposal status", str)
	}
}

// is defined ProposalType?
func validProposalStatus(status ProposalStatus) bool {
	if status == StatusDepositPeriod ||
		status == StatusVotingPeriod ||
		status == StatusPassed ||
		status == StatusRejected {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (status ProposalStatus) Marshal() ([]byte, error) {
	return []byte{byte(status)}, nil
}

// Unmarshal needed for protobuf compatibility
func (status *ProposalStatus) Unmarshal(data []byte) error {
	*status = ProposalStatus(data[0])
	return nil
}

// Marshals to JSON using string
func (status ProposalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (status *ProposalStatus) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := ProposalStatusFromString(s)
	if err != nil {
		return err
	}
	*status = bz2
	return nil
}

// Turns VoteStatus byte to String
func (status ProposalStatus) String() string {
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

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (status ProposalStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(status.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(status))))
	}
}

// Tally Results
type TallyResult struct {
	Yes        sdk.Int `json:"yes"`
	Abstain    sdk.Int `json:"abstain"`
	No         sdk.Int `json:"no"`
	NoWithVeto sdk.Int `json:"no_with_veto"`
}

func NewTallyResult(yes, abstain, no, noWithVeto sdk.Int) TallyResult {
	return TallyResult{
		Yes:        yes,
		Abstain:    abstain,
		No:         no,
		NoWithVeto: noWithVeto,
	}
}

func NewTallyResultFromMap(results map[VoteOption]sdk.Dec) TallyResult {
	return TallyResult{
		Yes:        results[OptionYes].TruncateInt(),
		Abstain:    results[OptionAbstain].TruncateInt(),
		No:         results[OptionNo].TruncateInt(),
		NoWithVeto: results[OptionNoWithVeto].TruncateInt(),
	}
}

// checks if two proposals are equal
func EmptyTallyResult() TallyResult {
	return TallyResult{
		Yes:        sdk.ZeroInt(),
		Abstain:    sdk.ZeroInt(),
		No:         sdk.ZeroInt(),
		NoWithVeto: sdk.ZeroInt(),
	}
}

// checks if two proposals are equal
func (tr TallyResult) Equals(comp TallyResult) bool {
	return (tr.Yes.Equal(comp.Yes) &&
		tr.Abstain.Equal(comp.Abstain) &&
		tr.No.Equal(comp.No) &&
		tr.NoWithVeto.Equal(comp.NoWithVeto))
}

func (tr TallyResult) String() string {
	return fmt.Sprintf(`Tally Result:
  Yes:        %s
  Abstain:    %s
  No:         %s
  NoWithVeto: %s`, tr.Yes, tr.Abstain, tr.No, tr.NoWithVeto)
}
