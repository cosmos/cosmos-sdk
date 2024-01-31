package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ GovHooks = MultiGovHooks{}

// combine multiple governance hooks, all hook functions are run in array sequence
type MultiGovHooks []GovHooks

func NewMultiGovHooks(hooks ...GovHooks) MultiGovHooks {
	return hooks
}

func (h MultiGovHooks) AfterProposalSubmission(ctx sdk.Context, proposalID uint64) error {
	var errs error
	for i := range h {
		errs = JoinErrors(errs, h[i].AfterProposalSubmission(ctx, proposalID))
	}
	return errs
}

func (h MultiGovHooks) AfterProposalDeposit(ctx sdk.Context, proposalID uint64, depositorAddr sdk.AccAddress) error {
	var errs error
	for i := range h {
		errs = JoinErrors(errs, h[i].AfterProposalDeposit(ctx, proposalID, depositorAddr))
	}
	return errs
}

func (h MultiGovHooks) AfterProposalVote(ctx sdk.Context, proposalID uint64, voterAddr sdk.AccAddress) error {
	var errs error
	for i := range h {
		errs = JoinErrors(errs, h[i].AfterProposalVote(ctx, proposalID, voterAddr))
	}
	return errs
}

func (h MultiGovHooks) AfterProposalFailedMinDeposit(ctx sdk.Context, proposalID uint64) error {
	var errs error
	for i := range h {
		errs = JoinErrors(errs, h[i].AfterProposalFailedMinDeposit(ctx, proposalID))
	}
	return errs
}

func (h MultiGovHooks) AfterProposalVotingPeriodEnded(ctx sdk.Context, proposalID uint64) error {
	var errs error
	for i := range h {
		errs = JoinErrors(errs, h[i].AfterProposalVotingPeriodEnded(ctx, proposalID))
	}
	return errs
}

// implementation of errors.Join() in Go 1.20, until we upgrade to that version
func JoinErrors(errs ...error) error {
	n := 0
	for _, err := range errs {
		if err != nil {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	e := &joinError{
		errs: make([]error, 0, n),
	}
	for _, err := range errs {
		if err != nil {
			e.errs = append(e.errs, err)
		}
	}
	return e
}

type joinError struct {
	errs []error
}

func (e *joinError) Error() string {
	var b []byte
	for i, err := range e.errs {
		if i > 0 {
			b = append(b, '\n')
		}
		b = append(b, err.Error()...)
	}
	return string(b)
}
