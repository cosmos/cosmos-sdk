package utils

import "github.com/pkg/errors"

// InvalidProposalID returns an error for an invalid Proposal ID
func InvalidProposalID(pid string) error {
	return errors.Errorf("proposal-id %s not a valid uint, please input a valid proposal-id", pid)
}

// FailedToFectchProposal returns an error for a failed store query
func FailedToFectchProposal(pid uint64, err error) error {
	return errors.Errorf("Failed to fetch proposal-id %d: %s", pid, err)
}
