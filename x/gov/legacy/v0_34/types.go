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

var (
	_ ProposalContent = TextProposal{}
	_ ProposalContent = SoftwareUpgradeProposal{}
)

const (
	ModuleName = "gov"

	StatusNil           ProposalStatus = 0x00
	StatusDepositPeriod ProposalStatus = 0x01
	StatusVotingPeriod  ProposalStatus = 0x02
	StatusPassed        ProposalStatus = 0x03
	StatusRejected      ProposalStatus = 0x04
	StatusFailed        ProposalStatus = 0x05

	OptionEmpty      VoteOption = 0x00
	OptionYes        VoteOption = 0x01
	OptionAbstain    VoteOption = 0x02
	OptionNo         VoteOption = 0x03
	OptionNoWithVeto VoteOption = 0x04

	ProposalTypeNil             ProposalKind = 0x00
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

type (
	SoftwareUpgradeProposal struct {
		TextProposal
	}

	ProposalQueue []uint64

	ProposalKind byte

	VoteOption     byte
	ProposalStatus byte

	ProposalContent interface {
		GetTitle() string
		GetDescription() string
		ProposalType() ProposalKind
	}

	Proposals []Proposal

	TextProposal struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	Proposal struct {
		ProposalContent `json:"proposal_content"`

		ProposalID uint64 `json:"proposal_id"`

		Status           ProposalStatus `json:"proposal_status"`
		FinalTallyResult TallyResult    `json:"final_tally_result"`

		SubmitTime     time.Time `json:"submit_time"`
		DepositEndTime time.Time `json:"deposit_end_time"`
		TotalDeposit   sdk.Coins `json:"total_deposit"`

		VotingStartTime time.Time `json:"voting_start_time"`
		VotingEndTime   time.Time `json:"voting_end_time"`
	}

	TallyParams struct {
		Quorum    sdk.Dec `json:"quorum,omitempty"`
		Threshold sdk.Dec `json:"threshold,omitempty"`
		Veto      sdk.Dec `json:"veto,omitempty"`
	}

	VotingParams struct {
		VotingPeriod time.Duration `json:"voting_period,omitempty"`
	}

	TallyResult struct {
		Yes        sdk.Int `json:"yes"`
		Abstain    sdk.Int `json:"abstain"`
		No         sdk.Int `json:"no"`
		NoWithVeto sdk.Int `json:"no_with_veto"`
	}

	Deposits []Deposit

	Vote struct {
		ProposalID uint64         `json:"proposal_id"`
		Voter      sdk.AccAddress `json:"voter"`
		Option     VoteOption     `json:"option"`
	}

	Votes []Vote

	DepositParams struct {
		MinDeposit       sdk.Coins     `json:"min_deposit,omitempty"`
		MaxDepositPeriod time.Duration `json:"max_deposit_period,omitempty"`
	}

	Deposit struct {
		ProposalID uint64         `json:"proposal_id"`
		Depositor  sdk.AccAddress `json:"depositor"`
		Amount     sdk.Coins      `json:"amount"`
	}

	DepositWithMetadata struct {
		ProposalID uint64  `json:"proposal_id"`
		Deposit    Deposit `json:"deposit"`
	}

	VoteWithMetadata struct {
		ProposalID uint64 `json:"proposal_id"`
		Vote       Vote   `json:"vote"`
	}

	GenesisState struct {
		StartingProposalID uint64                `json:"starting_proposal_id"`
		Deposits           []DepositWithMetadata `json:"deposits"`
		Votes              []VoteWithMetadata    `json:"votes"`
		Proposals          []Proposal            `json:"proposals"`
		DepositParams      DepositParams         `json:"deposit_params"`
		VotingParams       VotingParams          `json:"voting_params"`
		TallyParams        TallyParams           `json:"tally_params"`
	}
)

func (tp TextProposal) GetTitle() string           { return tp.Title }
func (tp TextProposal) GetDescription() string     { return tp.Description }
func (tp TextProposal) ProposalType() ProposalKind { return ProposalTypeText }

func (sup SoftwareUpgradeProposal) ProposalType() ProposalKind { return ProposalTypeSoftwareUpgrade }

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

func (status ProposalStatus) Marshal() ([]byte, error) {
	return []byte{byte(status)}, nil
}

func (status *ProposalStatus) Unmarshal(data []byte) error {
	*status = ProposalStatus(data[0])
	return nil
}

func (status ProposalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
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

func (vo VoteOption) Marshal() ([]byte, error) {
	return []byte{byte(vo)}, nil
}

func (vo *VoteOption) Unmarshal(data []byte) error {
	*vo = VoteOption(data[0])
	return nil
}

func (vo VoteOption) MarshalJSON() ([]byte, error) {
	return json.Marshal(vo.String())
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

func (pt ProposalKind) Marshal() ([]byte, error) {
	return []byte{byte(pt)}, nil
}

func (pt *ProposalKind) Unmarshal(data []byte) error {
	*pt = ProposalKind(data[0])
	return nil
}

func (pt ProposalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

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

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*ProposalContent)(nil), nil)
	cdc.RegisterConcrete(TextProposal{}, "gov/TextProposal", nil)
	cdc.RegisterConcrete(SoftwareUpgradeProposal{}, "gov/SoftwareUpgradeProposal", nil)
}
