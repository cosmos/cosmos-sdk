// DONTCOVER
// nolint
package v0_34

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const ModuleName = "gov"

const (
	ProposalTypeNil             ProposalKind = 0x00
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

const (
	OptionEmpty      VoteOption = 0x00
	OptionYes        VoteOption = 0x01
	OptionAbstain    VoteOption = 0x02
	OptionNo         VoteOption = 0x03
	OptionNoWithVeto VoteOption = 0x04
)

const (
	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
	StatusFailed        ProposalStatus = 0x05
)

type (
	GenesisState struct {
		StartingProposalID uint64                `json:"starting_proposal_id"`
		Deposits           []DepositWithMetadata `json:"deposits"`
		Votes              []VoteWithMetadata    `json:"votes"`
		Proposals          []Proposal            `json:"proposals"`
		DepositParams      DepositParams         `json:"deposit_params"`
		VotingParams       VotingParams          `json:"voting_params"`
		TallyParams        TallyParams           `json:"tally_params"`
	}

	DepositWithMetadata struct {
		ProposalID uint64  `json:"proposal_id"`
		Deposit    Deposit `json:"deposit"`
	}

	VoteWithMetadata struct {
		ProposalID uint64 `json:"proposal_id"`
		Vote       Vote   `json:"vote"`
	}

	Deposit struct {
		ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
		Depositor  sdk.AccAddress `json:"depositor"`   //  Address of the depositor
		Amount     sdk.Coins      `json:"amount"`      //  Deposit amount
	}

	Deposits []Deposit

	VoteOption byte

	Vote struct {
		ProposalID uint64         `json:"proposal_id"` //  proposalID of the proposal
		Voter      sdk.AccAddress `json:"voter"`       //  address of the voter
		Option     VoteOption     `json:"option"`      //  option from OptionSet chosen by the voter
	}

	Votes []Vote

	// Param around deposits for governance
	DepositParams struct {
		MinDeposit       sdk.Coins     `json:"min_deposit,omitempty"`        //  Minimum deposit for a proposal to enter voting period.
		MaxDepositPeriod time.Duration `json:"max_deposit_period,omitempty"` //  Maximum period for Atom holders to deposit on a proposal. Initial value: 2 months
	}

	TallyParams struct {
		Quorum    sdk.Dec `json:"quorum,omitempty"`    //  Minimum percentage of total stake needed to vote for a result to be considered valid
		Threshold sdk.Dec `json:"threshold,omitempty"` //  Minimum proportion of Yes votes for proposal to pass. Initial value: 0.5
		Veto      sdk.Dec `json:"veto,omitempty"`      //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
	}

	VotingParams struct {
		VotingPeriod time.Duration `json:"voting_period,omitempty"` //  Length of the voting period.
	}

	TallyResult struct {
		Yes        sdk.Int `json:"yes"`
		Abstain    sdk.Int `json:"abstain"`
		No         sdk.Int `json:"no"`
		NoWithVeto sdk.Int `json:"no_with_veto"`
	}

	ProposalKind byte

	ProposalStatus byte

	TextProposal struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	SoftwareUpgradeProposal struct {
		TextProposal
	}

	ProposalContent interface {
		GetTitle() string
		GetDescription() string
		ProposalType() ProposalKind
	}

	Proposal struct {
		ProposalContent `json:"proposal_content"` // Proposal content interface

		ProposalID uint64 `json:"proposal_id"` //  ID of the proposal

		Status           ProposalStatus `json:"proposal_status"`    //  Status of the Proposal {Pending, Active, Passed, Rejected}
		FinalTallyResult TallyResult    `json:"final_tally_result"` //  Result of Tallys

		SubmitTime     time.Time `json:"submit_time"`      //  Time of the block where TxGovSubmitProposal was included
		DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
		TotalDeposit   sdk.Coins `json:"total_deposit"`    //  Current deposit on this proposal. Initial value is set at InitialDeposit

		VotingStartTime time.Time `json:"voting_start_time"` //  Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
		VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
	}

	Proposals []Proposal
)

var _ ProposalContent = TextProposal{}
var _ ProposalContent = SoftwareUpgradeProposal{}

func (tp TextProposal) GetTitle() string       { return tp.Title }
func (tp TextProposal) GetDescription() string { return tp.Description }
func (tp TextProposal) ProposalType() ProposalKind        { return ProposalTypeText }

func (sup SoftwareUpgradeProposal) ProposalType() ProposalKind { return  ProposalTypeSoftwareUpgrade }


func (pt ProposalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (pt *ProposalKind) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	bz2, err := ProposalTypeFromString(s)
	if err != nil {
		return err
	}
	*pt = bz2
	return nil
}

func ProposalTypeFromString(str string) (ProposalKind, error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
	case "ParameterChange":
		return ProposalTypeParameterChange, nil
	case "SoftwareUpgrade":
		return ProposalTypeSoftwareUpgrade, nil
	default:
		return ProposalKind(0xff), fmt.Errorf("'%s' is not a valid proposal type", str)
	}
}

func (pt ProposalKind) String() string {
	switch pt {
	case ProposalTypeText:
		return "Text"
	case ProposalTypeParameterChange:
		return "ParameterChange"
	case ProposalTypeSoftwareUpgrade:
		return "SoftwareUpgrade"
	default:
		return ""
	}
}

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

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ProposalContent)(nil), nil)
	cdc.RegisterConcrete(TextProposal{}, "gov/TextProposal", nil)
}
