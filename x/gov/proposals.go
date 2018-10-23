package gov

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//-----------------------------------------------------------
// Proposal interface
type Proposal interface {
	GetProposalAbstract() ProposalAbstract
	Enact(sdk.Context, Keeper) error
}

// ProposalAbstract is a human-readable description about a proposal
type ProposalAbstract struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ProposalInfo is the status of a proposal managed by the keeper
type ProposalInfo struct {
	ProposalID int64

	Status      ProposalStatus //  Status of the Proposal {Pending, Active, Passed, Rejected}
	TallyResult TallyResult    //  Result of Tallys

	SubmitTime   time.Time //  Height of the block where TxGovSubmitProposal was included
	TotalDeposit sdk.Coins //  Current deposit on this proposal. Initial value is set at InitialDeposit

	VotingStartTime time.Time //  Height of the block where MinDeposit was reached. -1 if MinDeposit is not reached
}

func (info ProposalInfo) IsEmpty() bool {
	return info.ProposalID == 0
}

//-----------------------------------------------------------
// Text Proposals
type TextProposal struct {
	Abstract ProposalAbstract `json:"abstract"`

	ProposalType ProposalKind `json:"proposal_type"` //  Type of proposal. Initial set {PlainTextProposal, SoftwareUpgradeProposal}
}

func (tp TextProposal) GetProposalAbstract() ProposalAbstract { return tp.Abstract }
func (tp TextProposal) Enact(ctx sdk.Context, k Keeper) error { return nil }

// Implements Proposal Interface
var _ Proposal = TextProposal{}

// -------------------------------------------------------------
// Parameter Change Proposals

type ParameterChangeProposal struct {
	Abstract ProposalAbstract `json:"abstract"`

	StoreName string      `json:"store_name"`
	Key       []byte      `json:"key"`
	Subkey    []byte      `json:"subkey"`
	Value     interface{} `json:"value"`
}

func (pcp ParameterChangeProposal) GetProposalAbstract() ProposalAbstract { return pcp.Abstract }
func (pcp ParameterChangeProposal) Enact(ctx sdk.Context, k Keeper) error {
	s, ok := k.paramsKeeper.GetSubspace(pcp.StoreName)
	if !ok {
		return errors.New("Non-existing subspace")
	}
	if len(pcp.Subkey) == 0 {
		s.Set(ctx, pcp.Key, pcp.Value)
	} else {
		s.SetWithSubkey(ctx, pcp.Key, pcp.Subkey, pcp.Value)
	}
	return nil
}

var _ Proposal = ParameterChangeProposal{}

//-----------------------------------------------------------
// InfoQueue
type InfoQueue []int64

//-----------------------------------------------------------
// ProposalKind

// Type that represents Proposal Type as a byte
type ProposalKind byte

//nolint
const (
	ProposalTypeNil             ProposalKind = 0x00
	ProposalTypeText            ProposalKind = 0x01
	ProposalTypeParameterChange ProposalKind = 0x02
	ProposalTypeSoftwareUpgrade ProposalKind = 0x03
)

// String to proposalType byte.  Returns ff if invalid.
func ProposalTypeFromString(str string) (ProposalKind, error) {
	switch str {
	case "Text":
		return ProposalTypeText, nil
	case "ParameterChange":
		return ProposalTypeParameterChange, nil
	case "SoftwareUpgrade":
		return ProposalTypeSoftwareUpgrade, nil
	default:
		return ProposalKind(0xff), errors.Errorf("'%s' is not a valid proposal type", str)
	}
}

// is defined ProposalType?
func validProposalType(pt ProposalKind) bool {
	if pt == ProposalTypeText ||
		pt == ProposalTypeParameterChange ||
		pt == ProposalTypeSoftwareUpgrade {
		return true
	}
	return false
}

// Marshal needed for protobuf compatibility
func (pt ProposalKind) Marshal() ([]byte, error) {
	return []byte{byte(pt)}, nil
}

// Unmarshal needed for protobuf compatibility
func (pt *ProposalKind) Unmarshal(data []byte) error {
	*pt = ProposalKind(data[0])
	return nil
}

// Marshals to JSON using string
func (pt ProposalKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.String())
}

// Unmarshals from JSON assuming Bech32 encoding
func (pt *ProposalKind) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return nil
	}

	bz2, err := ProposalTypeFromString(s)
	if err != nil {
		return err
	}
	*pt = bz2
	return nil
}

// Turns VoteOption byte to String
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

// For Printf / Sprintf, returns bech32 when using %s
// nolint: errcheck
func (pt ProposalKind) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(pt.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(pt))))
	}
}

//-----------------------------------------------------------
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
		return ProposalStatus(0xff), errors.Errorf("'%s' is not a valid proposal status", str)
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
		return nil
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

//-----------------------------------------------------------
// Tally Results
type TallyResult struct {
	Yes        sdk.Dec `json:"yes"`
	Abstain    sdk.Dec `json:"abstain"`
	No         sdk.Dec `json:"no"`
	NoWithVeto sdk.Dec `json:"no_with_veto"`
}

// checks if two proposals are equal
func EmptyTallyResult() TallyResult {
	return TallyResult{
		Yes:        sdk.ZeroDec(),
		Abstain:    sdk.ZeroDec(),
		No:         sdk.ZeroDec(),
		NoWithVeto: sdk.ZeroDec(),
	}
}

// checks if two proposals are equal
func (resultA TallyResult) Equals(resultB TallyResult) bool {
	return (resultA.Yes.Equal(resultB.Yes) &&
		resultA.Abstain.Equal(resultB.Abstain) &&
		resultA.No.Equal(resultB.No) &&
		resultA.NoWithVeto.Equal(resultB.NoWithVeto))
}
