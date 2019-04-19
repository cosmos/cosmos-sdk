package legacy

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var cdc *codec.Codec

func init() {
	cdc = codec.New()
	cdc.RegisterConcrete(Proposal{}, "gov/TextProposal", nil)
	cdc.Seal()
}

type ProposalKind byte

const (
	ProposalTypeNil             ProposalKind = 0x00
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

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

func ProposalTypeFromString(str string) (ProposalKind, error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
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
	case ProposalTypeSoftwareUpgrade:
		return "SoftwareUpgrade"
	default:
		return ""
	}
}

type Proposal struct {
	ProposalID   uint64       `json:"proposal_id"`   //  ID of the proposal
	Title        string       `json:"title"`         //  Title of the proposal
	Description  string       `json:"description"`   //  Description of the proposal
	ProposalType ProposalKind `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}

	Status           gov.ProposalStatus `json:"proposal_status"`    //  Status of the Proposal {Pending, Active, Passed, Rejected}
	FinalTallyResult gov.TallyResult    `json:"final_tally_result"` //  Result of Tallys

	SubmitTime     time.Time `json:"submit_time"`      //  Time of the block where TxGovSubmitProposal was included
	DepositEndTime time.Time `json:"deposit_end_time"` // Time that the Proposal would expire if deposit amount isn't met
	TotalDeposit   sdk.Coins `json:"total_deposit"`    //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time `json:"voting_start_time"` //  Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	VotingEndTime   time.Time `json:"voting_end_time"`   // Time that the VotingPeriod for this proposal will end and votes will be tallied
}

func (p Proposal) Migrate() (res gov.Proposal, err error) {
	content := gov.ContentFromProposalType(p.Title, p.Description, p.ProposalType.String())
	if content == nil {
		err = fmt.Errorf("invalid proposal kind %v", p.ProposalType)
		return
	}
	res = gov.Proposal{
		Content:          content,
		ProposalID:       p.ProposalID,
		Status:           p.Status,
		FinalTallyResult: p.FinalTallyResult,
		SubmitTime:       p.SubmitTime,
		DepositEndTime:   p.DepositEndTime,
		TotalDeposit:     p.TotalDeposit,
		VotingStartTime:  p.VotingStartTime,
		VotingEndTime:    p.VotingEndTime,
	}
	return
}
