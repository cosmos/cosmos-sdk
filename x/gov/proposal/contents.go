package proposal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/errors"
)

// constant maximum size for content abstract
const (
	MaxDescriptionLength int = 5000
	MaxTitleLength       int = 140
)

// Checks length of title and description
func IsValidContent(codespace sdk.CodespaceType, title, description string) sdk.Error {
	if len(title) == 0 {
		return errors.ErrInvalidTitle(codespace, "No title present in proposal")
	}
	if len(title) > MaxTitleLength {
		return errors.ErrInvalidTitle(codespace, fmt.Sprintf("Proposal title is longer than max length of %d", MaxTitleLength))
	}
	if len(description) == 0 {
		return errors.ErrInvalidDescription(codespace, "No description present in proposal")
	}
	if len(description) > MaxDescriptionLength {
		return errors.ErrInvalidDescription(codespace, fmt.Sprintf("Proposal description is longer than max length of %d", MaxDescriptionLength))
	}
	return nil
}

// Content is an interface that has title, description, and proposaltype
// that the governance module can use to identify them and generate human readable messages
// Content can have additional fields, which will handled by ProposalHandlers
// via type assertion, e.g. parameter change amount in ParameterChangeProposal
type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
}

// Text Proposals
type Abstract struct {
	Title       string `json:"title"`       //  Title of the proposal
	Description string `json:"description"` //  Description of the proposal
}

func NewAbstract(title, description string) Abstract {
	return Abstract{
		Title:       title,
		Description: description,
	}
}

// nolint
func (abs Abstract) GetTitle() string       { return abs.Title }
func (abs Abstract) GetDescription() string { return abs.Description }

// Handler handles the proposals after it has passed the governance process
type Handler func(ctx sdk.Context, content Content) sdk.Error

// Proto is used to generate content from SubmitForm
type Proto func(title, description string) Content
