package v1beta2

import sdk "github.com/cosmos/cosmos-sdk/types"

// Copied over from /x/gov/types/keys.go to avoid circular imports
const (
	ModuleName = "gov"

	RouterKey = ModuleName
)

// Content defines an interface that a proposal must implement. It contains
// information such as the title and description along with the type and routing
// information for the appropriate handler to process the proposal. Content can
// have additional fields, which will handled by a proposal's Handler.
type Content interface {
	GetTitle() string
	GetDescription() string
	ProposalRoute() string
	ProposalType() string
	ValidateBasic() error
	String() string
}

// Handler defines a function that handles a proposal after it has passed the
// governance process.
type Handler func(ctx sdk.Context, content Content) error
