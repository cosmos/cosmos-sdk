package gov

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
	StartingProposalID uint64                `json:"starting_proposal_id"`
	Deposits           []DepositWithMetadata `json:"deposits"`
	Votes              []VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal            `json:"proposals"`
	DepositParams      DepositParams         `json:"deposit_params"`
	VotingParams       VotingParams          `json:"voting_params"`
	TallyParams        TallyParams           `json:"tally_params"`
}

type DepositWithMetadata struct {
	ProposalID uint64  `json:"proposal_id"`
	Deposit    Deposit `json:"deposit"`
}

type VoteWithMetadata struct {
	ProposalID uint64 `json:"proposal_id"`
	Vote       Vote   `json:"vote"`
}

type Proposal struct {
	Content `json:"content"` // Proposal content interface

	ProposalID       uint64         `json:"proposal_id"`        //  ID of the proposal
	Status           ProposalStatus `json:"proposal_status"`    // Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult TallyResult    `json:"final_tally_result"` // Result of Tallys

	SubmitTime     time.Time `json:"submit_time"`      // Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins `json:"total_deposit"`    // Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` // Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

type Proposals []Proposal

type Deposit struct {
	ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
	Depositor  sdk.AccAddress `json:"depositor"`   //  Address of the depositor
	Amount     sdk.Coins      `json:"amount"`      //  Deposit amount
}

type Deposits []Deposit

type Vote struct {
	ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
	Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
	Option     VoteOption     `json:"option"`      //  option from OptionSet chosen by the voter
}

type Votes []Vote

// Param around deposits for governance
type DepositParams struct {
	MinDeposit       sdk.Coins     `json:"min_deposit,omitempty"`        //  Minimum deposit for a proposal to enter voting period.
	MaxDepositPeriod time.Duration `json:"max_deposit_period,omitempty"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
}

type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() sdk.Error
	String() string
}

type TallyParams struct {
	Quorum    sdk.Dec `json:"quorum,omitempty"`    //  Minimum percentage of total stake needed to vote for a result to be considered valid
	Threshold sdk.Dec `json:"threshold,omitempty"` //  Minimum proportion of Yes votes for proposal to pass. Initial value: 0.5
	Veto      sdk.Dec `json:"veto,omitempty"`      //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
}

type VotingParams struct {
	VotingPeriod time.Duration `json:"voting_period,omitempty"` //  Length of the voting period.
}

type TallyResult struct {
	Yes        sdk.Int `json:"yes"`
	Abstain    sdk.Int `json:"abstain"`
	No         sdk.Int `json:"no"`
	NoWithVeto sdk.Int `json:"no_with_veto"`
}

// ----------------------------------------------------------------------------
// ProposalStatus
// ----------------------------------------------------------------------------

// ProposalStatus is a type alias that represents a proposal status as a byte
type ProposalStatus byte

//nolint
const (
	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
	StatusFailed        ProposalStatus = 0x05
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

	case "Failed":
		return StatusFailed, nil

	case "":
		return StatusNil, nil

	default:
		return ProposalStatus(0xff), fmt.Errorf("'%s' is not a valid proposal status", str)
	}
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

// String implements the Stringer interface.
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

	case StatusFailed:
		return "Failed"

	default:
		return ""
	}
}

// Format implements the fmt.Formatter interface.
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

// ----------------------------------------------------------------------------
// VoteOption
// ----------------------------------------------------------------------------

type VoteOption byte

// Vote options
const (
	OptionEmpty      VoteOption = 0x00
	OptionYes        VoteOption = 0x01
	OptionAbstain    VoteOption = 0x02
	OptionNo         VoteOption = 0x03
	OptionNoWithVeto VoteOption = 0x04
)

// VoteOptionFromString returns a VoteOption from a string. It returns an error
// if the string is invalid.
func VoteOptionFromString(str string) (VoteOption, error) {
	switch str {
	case "Yes":
		return OptionYes, nil

	case "Abstain":
		return OptionAbstain, nil

	case "No":
		return OptionNo, nil

	case "NoWithVeto":
		return OptionNoWithVeto, nil

	default:
		return VoteOption(0xff), fmt.Errorf("'%s' is not a valid vote option", str)
	}
}

// Marshal needed for protobuf compatibility.
func (vo VoteOption) Marshal() ([]byte, error) {
	return []byte{byte(vo)}, nil
}

// Unmarshal needed for protobuf compatibility.
func (vo *VoteOption) Unmarshal(data []byte) error {
	*vo = VoteOption(data[0])
	return nil
}

// Marshals to JSON using string.
func (vo VoteOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(vo.String())
}

// UnmarshalJSON decodes from JSON assuming Bech32 encoding.
func (vo *VoteOption) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := VoteOptionFromString(s)
	if err != nil {
		return err
	}

	*vo = bz2
	return nil
}

// String implements the Stringer interface.
func (vo VoteOption) String() string {
	switch vo {
	case OptionYes:
		return "Yes"
	case OptionAbstain:
		return "Abstain"
	case OptionNo:
		return "No"
	case OptionNoWithVeto:
		return "NoWithVeto"
	default:
		return ""
	}
}

// Format implements the fmt.Formatter interface.
// nolint: errcheck
func (vo VoteOption) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(vo.String()))
	default:
		s.Write([]byte(fmt.Sprintf("%v", byte(vo))))
	}
}

// ----------------------------------------------------------------------------
// Codec
// ----------------------------------------------------------------------------

func RegisterCodec(codec *codec.Codec) {
	codec.RegisterInterface((*Content)(nil), nil)
}
