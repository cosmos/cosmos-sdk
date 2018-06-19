package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	errValidatorsChanged = fmt.Errorf("Validators differ between header and certifier")
	errCommitNotFound    = fmt.Errorf("Commit not found by provider")
	errTooMuchChange     = fmt.Errorf("Validators change too much to safely update")
	errPastTime          = fmt.Errorf("Update older than certifier height")
	errNoPathFound       = fmt.Errorf("Cannot find a path of validators")
)

// IsCommitNotFoundErr checks whether an error is due to missing data
func IsCommitNotFoundErr(err error) bool {
	return err != nil && (errors.Cause(err) == errCommitNotFound)
}

// ErrCommitNotFound indicates that a the requested commit was not found.
func ErrCommitNotFound() error {
	return errors.WithStack(errCommitNotFound)
}

// IsValidatorsChangedErr checks whether an error is due
// to a differing validator set.
func IsValidatorsChangedErr(err error) bool {
	return err != nil && (errors.Cause(err) == errValidatorsChanged)
}

// ErrValidatorsChanged indicates that the validator set was changed between two commits.
func ErrValidatorsChanged() error {
	return errors.WithStack(errValidatorsChanged)
}

// IsTooMuchChangeErr checks whether an error is due to too much change
// between these validators sets.
func IsTooMuchChangeErr(err error) bool {
	return err != nil && (errors.Cause(err) == errTooMuchChange)
}

// ErrTooMuchChange indicates that the underlying validator set was changed by >1/3.
func ErrTooMuchChange() error {
	return errors.WithStack(errTooMuchChange)
}

// IsPastTimeErr ...
func IsPastTimeErr(err error) bool {
	return err != nil && (errors.Cause(err) == errPastTime)
}

// ErrPastTime ...
func ErrPastTime() error {
	return errors.WithStack(errPastTime)
}

// IsNoPathFoundErr checks whether an error is due to no path of
// validators in provider from where we are to where we want to be
func IsNoPathFoundErr(err error) bool {
	return err != nil && (errors.Cause(err) == errNoPathFound)
}

// ErrNoPathFound ...
func ErrNoPathFound() error {
	return errors.WithStack(errNoPathFound)
}

//--------------------------------------------

type errHeightMismatch struct {
	h1, h2 int64
}

func (e errHeightMismatch) Error() string {
	return fmt.Sprintf("Blocks don't match - %d vs %d", e.h1, e.h2)
}

// IsHeightMismatchErr checks whether an error is due to data from different blocks
func IsHeightMismatchErr(err error) bool {
	if err == nil {
		return false
	}
	_, ok := errors.Cause(err).(errHeightMismatch)
	return ok
}

// ErrHeightMismatch returns an mismatch error with stack-trace
func ErrHeightMismatch(h1, h2 int64) error {
	return errors.WithStack(errHeightMismatch{h1, h2})
}
