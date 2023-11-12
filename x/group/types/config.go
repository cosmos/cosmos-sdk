package types

import "time"

// Config used to initialize x/group module avoiding using global variable.
type Config struct {
	// MaxExecutionPeriod defines the max duration after a proposal's voting
	// period ends that members can send a MsgExec to execute the proposal.
	MaxExecutionPeriod time.Duration

	// MaxMetadataLen defines the max chars allowed in all
	// messages that allows creating or updating a group
	// with a metadata field
	MaxMetadataLen uint64

	// MaxProposalTitleLen defines the max chars allowed
	// in string for the MsgSubmitProposal and Proposal
	// summary field
	MaxProposalTitleLen uint64

	// MaxProposalSummaryLen defines the max chars allowed
	// in string for the MsgSubmitProposal and Proposal
	// summary field
	MaxProposalSummaryLen uint64
}

// DefaultConfig returns the default config for group.
func DefaultConfig() Config {
	return Config{
		MaxExecutionPeriod:    2 * time.Hour * 24 * 7, // Two weeks.
		MaxMetadataLen:        255,
		MaxProposalTitleLen:   255,
		MaxProposalSummaryLen: 10200,
	}
}
