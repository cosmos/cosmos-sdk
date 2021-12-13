package types

import (
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
)

// DefaultStartingProposalID is 1
const DefaultStartingProposalID uint64 = 1

// NewProposal creates a new Proposal instance
func NewProposal(messages []sdk.Msg, id uint64, submitTime, depositEndTime time.Time) (Proposal, error) {

	msgs, err := sdktx.SetMsgs(messages)
	if err != nil {
		return Proposal{}, err
	}

	p := Proposal{
		ProposalId:       id,
		Messages:         msgs,
		Status:           StatusDepositPeriod,
		FinalTallyResult: EmptyTallyResult(),
		TotalDeposit:     sdk.NewCoins(),
		SubmitTime:       submitTime,
		DepositEndTime:   depositEndTime,
	}

	return p, nil
}

// String implements stringer interface
func (p Proposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// GetMessages returns the proposal messages
func (p Proposal) GetMessages() ([]sdk.Msg, error) {
	return sdktx.GetMsgs(p.Messages, "sdk.MsgProposal")
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return sdktx.UnpackInterfaces(unpacker, p.Messages)
}

// Proposals is an array of proposal
type Proposals []Proposal

var _ types.UnpackInterfacesMessage = Proposals{}

// Equal returns true if two slices (order-dependant) of proposals are equal.
func (p Proposals) Equal(other Proposals) bool {
	if len(p) != len(other) {
		return false
	}

	for i, proposal := range p {
		if !proposal.Equal(other[i]) {
			return false
		}
	}

	return true
}

// String implements stringer interface
func (p Proposals) String() string {
	out := "ID - (Status) [Type] Title\n"
	for _, prop := range p {
		out += fmt.Sprintf("%d - %s\n",
			prop.ProposalId, prop.Status)
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
