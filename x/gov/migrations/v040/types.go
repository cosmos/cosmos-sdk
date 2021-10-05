package v040

import (
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

type (
	Content interface {
		GetTitle() string
		GetDescription() string
		ProposalRoute() string
		ProposalType() string
		ValidateBasic() error
		String() string
	}

	Proposal struct {
		ProposalId       uint64
		Content          Content
		Status           types.ProposalStatus
		FinalTallyResult types.TallyResult
		SubmitTime       time.Time
		DepositEndTime   time.Time
		TotalDeposit     sdk.Coins
		VotingStartTime  time.Time
		VotingEndTime    time.Time
	}

	Proposals []Proposal

	GenesisState struct {
		StartingProposalId uint64
		Deposits           types.Deposits
		Votes              types.Votes
		Proposals          Proposals
		DepositParams      types.DepositParams
		VotingParams       types.VotingParams
		TallyParams        types.TallyParams
	}

	TextProposal struct {
		Title       string
		Description string
	}
)

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
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal title cannot be blank")
	}
	if len(title) > MaxTitleLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal title is longer than max length of %d", MaxTitleLength)
	}

	description := c.GetDescription()
	if len(description) == 0 {
		return sdkerrors.Wrap(types.ErrInvalidProposalContent, "proposal description cannot be blank")
	}
	if len(description) > MaxDescriptionLength {
		return sdkerrors.Wrapf(types.ErrInvalidProposalContent, "proposal description is longer than max length of %d", MaxDescriptionLength)
	}

	return nil
}
