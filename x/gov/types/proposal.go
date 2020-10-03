package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	yaml "gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultStartingProposalID is 1
const DefaultStartingProposalID uint64 = 1

// NewProposal creates a new Proposal instance
func NewProposal(content Content, id uint64, submitTime, depositEndTime time.Time) (Proposal, error) {
	p := Proposal{
		ProposalId:       id,
		Status:           StatusDepositPeriod,
		FinalTallyResult: EmptyTallyResult(),
		TotalDeposit:     sdk.NewCoins(),
		SubmitTime:       submitTime,
		DepositEndTime:   depositEndTime,
	}

	msg, ok := content.(proto.Message)
	if !ok {
		return Proposal{}, fmt.Errorf("%T does not implement proto.Message", content)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return Proposal{}, err
	}

	p.Content = any

	return p, nil
}

// String implements stringer interface
func (p Proposal) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// GetContent returns the proposal Content
func (p Proposal) GetContent() Content {
	content, ok := p.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

func (p Proposal) ProposalType() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.ProposalType()
}

func (p Proposal) ProposalRoute() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.ProposalRoute()
}

func (p Proposal) GetTitle() string {
	content := p.GetContent()
	if content == nil {
		return ""
	}
	return content.GetTitle()
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (p Proposal) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	var content Content
	return unpacker.UnpackAny(p.Content, &content)
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
		out += fmt.Sprintf("%d - (%s) [%s] %s\n",
			prop.ProposalId, prop.Status,
			prop.ProposalType(), prop.GetTitle())
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

// Implements Content Interface
var _ Content = &TextProposal{}

// NewTextProposal creates a text proposal Content
func NewTextProposal(title, description string) Content {
	return &TextProposal{title, description}
}

// GetTitle returns the proposal title
func (tp *TextProposal) GetTitle() string { return tp.Title }

// GetDescription returns the proposal description
func (tp *TextProposal) GetDescription() string { return tp.Description }

// ProposalRoute returns the proposal router key
func (tp *TextProposal) ProposalRoute() string { return RouterKey }

// ProposalType is "Text"
func (tp *TextProposal) ProposalType() string { return ProposalTypeText }

// ValidateBasic validates the content's title and description of the proposal
func (tp *TextProposal) ValidateBasic() error { return ValidateAbstract(tp) }

// String implements Stringer interface
func (tp TextProposal) String() string {
	out, _ := yaml.Marshal(tp)
	return string(out)
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
