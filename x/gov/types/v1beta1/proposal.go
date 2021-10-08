package v1beta1

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewProposal(content Content, id uint64, submitTime, depositEndTime time.Time) (Proposal, error) {
	msg, ok := content.(proto.Message)
	if !ok {
		return Proposal{}, fmt.Errorf("%T does not implement proto.Message", content)
	}

	any, err := types.NewAnyWithValue(msg)
	if err != nil {
		return Proposal{}, err
	}

	p := Proposal{
		Content:          any,
		ProposalId:       id,
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

// GetContent returns the proposal Content
func (p Proposal) GetContent() Content {
	content, ok := p.Content.GetCachedValue().(Content)
	if !ok {
		return nil
	}
	return content
}

// Proposals is an array of proposal
type Proposals []Proposal

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

// Proposal types
const (
	ProposalTypeText string = "Text"

	// Constants pertaining to a Content object
	MaxDescriptionLength int = 5000
	MaxTitleLength       int = 140
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

// ValidateAbstract validates a proposal's abstract contents returning an error
// if invalid.
func ValidateAbstract(c Content) error {
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return sdkerrors.Wrap(ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return sdkerrors.Wrapf(ErrInvalidProposalContent, "proposal title is longer than max length of %d", MaxTitleLength)
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return sdkerrors.Wrap(ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return sdkerrors.Wrapf(ErrInvalidProposalContent, "proposal description is longer than max length of %d", MaxDescriptionLength)
	}

	return nil
}
