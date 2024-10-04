package v1

import (
	"context"
)

// TallyHandler is the interface that is used for tallying votes.
type TallyHandler interface {
	Tally(context.Context, Proposal) (passes bool, burnDeposits bool, tallyResults TallyResult, err error)
}
