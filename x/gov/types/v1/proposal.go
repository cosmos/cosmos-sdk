package v1

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

const (
	// DefaultStartingProposalID is 1
	DefaultStartingProposalID uint64 = 1

	StatusNil           = ProposalStatus_PROPOSAL_STATUS_UNSPECIFIED
	StatusDepositPeriod = ProposalStatus_PROPOSAL_STATUS_DEPOSIT_PERIOD
	StatusVotingPeriod  = ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD
	StatusPassed        = ProposalStatus_PROPOSAL_STATUS_PASSED
	StatusRejected      = ProposalStatus_PROPOSAL_STATUS_REJECTED
	StatusFailed        = ProposalStatus_PROPOSAL_STATUS_FAILED
)

// NewProposal creates a new Proposal instance
func NewProposal(messages []sdk.Msg, id uint64, metadata string, submitTime, depositEndTime time.Time) (Proposal, error) {
	msgs, err := sdktx.SetMsgs(messages)
	if err != nil {
		return Proposal{}, err
	}

	tally := EmptyTallyResult()

	p := Proposal{
		Id:               id,
		Messages:         msgs,
		Metadata:         metadata,
		Status:           StatusDepositPeriod,
		FinalTallyResult: &tally,
		SubmitTime:       &submitTime,
		DepositEndTime:   &depositEndTime,
	}

	return p, nil
}

// GetMessages returns the proposal messages
func (p Proposal) GetMsgs() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(p.Messages, "sdk.MsgProposal")
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, p.Messages)
}

// Proposals is an array of proposal
type Proposals []*Proposal

var _ types.UnpackInterfacesMessage = Proposals{}

// String implements stringer interface
func (p Proposals) String() string {
	out := "ID - (Status) [Type] Title\n"
	for _, prop := range p {
		out += fmt.Sprintf("%d - %s\n",
			prop.Id, prop.Status)
	}
	return strings.TrimSpace(out)
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposals) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, x := range p {
		err := x.UnpackInterfaces(unpacker)
		if err != nil {
			return err
		}
	}
	return nil
}

type (
	// ProposalQueue defines a queue for proposal ids
	ProposalQueue []uint64
)

// ProposalStatusFromString turns a string into a ProposalStatus
func ProposalStatusFromString(str string) (ProposalStatus, error) {
	num, ok := ProposalStatus_value[str]
	if !ok {
		return StatusNil, fmt.Errorf("'%s' is not a valid proposal status", str)
	}
	return ProposalStatus(num), nil
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

// Format implements the fmt.Formatter interface.

func (status ProposalStatus) Format(s fmt.State, verb rune) {
	switch verb {
	case 's':
		s.Write([]byte(status.String()))
	default:
		// TODO: Do this conversion more directly
		s.Write([]byte(fmt.Sprintf("%v", byte(status))))
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
