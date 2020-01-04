package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultStartingProposalID is 1
const DefaultStartingProposalID uint64 = 1

// Proposal defines an interface used by the governance module to allow for voting
// on network changes.
type Proposal interface {
	sdk.Marshaler

	GetContent() Content               // Proposal content interface
	SetContent(content Content) error

	GetProposalID() uint64             //  ID of the proposal
	SetProposalID(id uint64)

	GetStatus() ProposalStatus         // Status of the Proposal {Pending, Active, Passed, Rejected}
	SetStatus(status ProposalStatus)

	GetFinalTallyResult() *TallyResult // Result of Tallys
	SetFinalTallyResult(result *TallyResult)

	GetSubmitTime() time.Time          // Time of the block where TxGovSubmitProposal was included
	SetSubmitTime(time time.Time)

	GetDepositEndTime() time.Time      // Time that the Proposal would expire if deposit amount isn't met
	SetDepositEndTime(time time.Time)

	GetTotalDeposit() sdk.Coins        // Current deposit on this proposal. Initial value is set at InitialDeposit
	SetTotalDeposit(coins sdk.Coins)

	GetVotingStartTime() time.Time     // Time of the block where MinDeposit was reached. -1 if MinDeposit is not reached
	SetVotingStartTime(time.Time)

	GetVotingEndTime() time.Time       // Time that the VotingPeriod for this proposal will end and votes will be tallied
	SetVotingEndTime(time.Time)
}

func (m ProposalBase) GetProposalID() uint64 {
	return m.ProposalID
}

func (m ProposalBase) GetStatus() ProposalStatus {
	return m.Status
}

func (m *ProposalBase) GetFinalTallyResult() *TallyResult {
	return m.FinalTallyResult
}

func (m *ProposalBase) GetSubmitTime() time.Time {
	return m.GetSubmitTime()
}

func (m *ProposalBase) GetDepositEndTime() time.Time {
	return sdk.ProtoTimestampToTime(m.DepositEndTime)
}

func (m *ProposalBase) GetTotalDeposit() sdk.Coins {
	return m.TotalDeposit
}

func (m *ProposalBase) GetVotingStartTime() time.Time {
	return sdk.ProtoTimestampToTime(m.VotingStartTime)
}

func (m *ProposalBase) GetVotingEndTime() time.Time {
	return sdk.ProtoTimestampToTime(m.VotingEndTime)
}

func (m *ProposalBase) SetProposalID(id uint64) {
	m.ProposalID = id
}

func (m *ProposalBase) SetStatus(status ProposalStatus) {
	m.Status = status
}

func (m *ProposalBase) SetFinalTallyResult(result *TallyResult) {
	m.FinalTallyResult = result
}

func (m *ProposalBase) SetSubmitTime(t time.Time) {
	m.SubmitTime = sdk.TimeToProtoTimestamp(t)
}

func (m *ProposalBase) SetDepositEndTime(t time.Time) {
	m.DepositEndTime = sdk.TimeToProtoTimestamp(t)
}

func (m *ProposalBase) SetTotalDeposit(coins sdk.Coins) {
	m.TotalDeposit = coins
}

func (m *ProposalBase) SetVotingStartTime(t time.Time) {
	m.VotingStartTime = sdk.TimeToProtoTimestamp(t)
}

func (m *ProposalBase) SetVotingEndTime(t time.Time) {
	m.VotingEndTime = sdk.TimeToProtoTimestamp(t)
}

var _ Proposal = &BasicProposal{}

func (m *BasicProposal) GetContent() Content {
	return m.Content.GetContent()
}

func (m *BasicProposal) SetContent(value Content) error {
	return m.Content.SetContent(value)
}

func ProposalToString(p Proposal) string {
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
		p.GetProposalID(), p.GetContent().GetTitle(), p.GetContent().ProposalType(),
		p.GetStatus(), p.GetSubmitTime(), p.GetDepositEndTime(),
		p.GetTotalDeposit(), p.GetVotingStartTime(), p.GetVotingEndTime(), p.GetContent().GetDescription(),
	)
}

// Proposals is an array of proposal
type Proposals []Proposal

// String implements stringer interface
func (p Proposals) String() string {
	out := "ID - (Status) [Type] Title\n"
	for _, prop := range p {
		out += fmt.Sprintf("%d - (%s) [%s] %s\n",
			prop.GetProposalID(), prop.GetStatus(),
			prop.GetContent().ProposalType(), prop.GetContent().GetTitle())
	}
	return strings.TrimSpace(out)
}

type (
	// ProposalQueue defines a queue for proposal ids
	ProposalQueue []uint64
)

// ProposalStatusFromString turns a string into a ProposalStatus
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

// ValidProposalStatus returns true if the proposal status is valid and false
// otherwise.
func ValidProposalStatus(status ProposalStatus) bool {
	if status == StatusDepositPeriod ||
		status == StatusVotingPeriod ||
		status == StatusPassed ||
		status == StatusRejected ||
		status == StatusFailed {
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

// MarshalJSON Marshals to JSON using string representation of the status
func (status ProposalStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// UnmarshalJSON Unmarshals from JSON assuming Bech32 encoding
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

// Proposal types
const (
	ProposalTypeText string = "Text"
)

// NewTextProposal creates a text proposal Content
func NewTextProposal(title, description string) Content {
	return TextProposal{title, description}
}

// Implements Content Interface
var _ Content = TextProposal{}

// GetTitle returns the proposal title
func (tp TextProposal) GetTitle() string { return tp.Title }

// GetDescription returns the proposal description
func (tp TextProposal) GetDescription() string { return tp.Description }

// ProposalRoute returns the proposal router key
func (tp TextProposal) ProposalRoute() string { return RouterKey }

// ProposalType is "Text"
func (tp TextProposal) ProposalType() string { return ProposalTypeText }

// ValidateBasic validates the content's title and description of the proposal
func (tp TextProposal) ValidateBasic() error { return ValidateAbstract(tp) }

// String implements Stringer interface
func (tp TextProposal) String() string {
	return fmt.Sprintf(`Text Proposal:
  Title:       %s
  Description: %s
`, tp.Title, tp.Description)
}

var validProposalTypes = map[string]struct{}{
	ProposalTypeText: {},
}

// RegisterProposalType registers a proposal type. It will panic if the type is
// already registered.
func RegisterProposalType(ty string) {
	if _, ok := validProposalTypes[ty]; ok {
		panic(fmt.Sprintf("already registered proposal type: %s", ty))
	}

	validProposalTypes[ty] = struct{}{}
}

// ContentFromProposalType returns a Content object based on the proposal type.
func ContentFromProposalType(title, desc, ty string) Content {
	switch ty {
	case ProposalTypeText:
		return NewTextProposal(title, desc)

	default:
		return nil
	}
}

// IsValidProposalType returns a boolean determining if the proposal type is
// valid.
//
// NOTE: Modules with their own proposal types must register them.
func IsValidProposalType(ty string) bool {
	_, ok := validProposalTypes[ty]
	return ok
}

// ProposalHandler implements the Handler interface for governance module-based
// proposals (ie. TextProposal ). Since these are
// merely signaling mechanisms at the moment and do not affect state, it
// performs a no-op.
func ProposalHandler(_ sdk.Context, c Content) error {
	switch c.ProposalType() {
	case ProposalTypeText:
		// both proposal types do not change state so this performs a no-op
		return nil

	default:
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized gov proposal type: %s", c.ProposalType())
	}
}
