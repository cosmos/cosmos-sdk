package proposal

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/errors"
)

// constant maximum size for content abstract
const (
	MaxDescriptionLength int = 5000
	MaxTitleLength       int = 140
)

// Checks length of title and description
func ValidateAbstract(codespace sdk.CodespaceType, c Content) sdk.Error {
	title := c.GetTitle()
	if len(strings.TrimSpace(title)) == 0 {
		return errors.ErrInvalidTitle(codespace, "No title present in proposal")
	}
	if len(title) > MaxTitleLength {
		return errors.ErrInvalidTitle(codespace, fmt.Sprintf("Proposal title is longer than max length of %d", MaxTitleLength))
	}
	description := c.GetDescription()
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
	ValidateBasic() sdk.Error
	String() string
}

// Handler handles the proposals after it has passed the governance process
type Handler func(ctx sdk.Context, content Content) sdk.Error
