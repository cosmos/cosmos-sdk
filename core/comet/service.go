package comet

import (
	"context"

	"cosmossdk.io/core/abci"
)

// Service is an interface that can be used to get information specific to Comet
type Service interface {
	CometInfo(context.Context) Info
}

// Info is the information comet provides apps in ABCI
type Info abci.Info

// MisbehaviorType is the type of misbehavior for a validator
type MisbehaviorType abci.MisbehaviorType

const (
	Unknown           = abci.Unknown
	DuplicateVote     = abci.DuplicateVote
	LightClientAttack = abci.LightClientAttack
)

// Evidence is the misbehavior information of ABCI
type Evidence abci.Evidence

// CommitInfo is the commit information of ABCI
type CommitInfo abci.CommitInfo

// VoteInfo is the vote information of ABCI
type VoteInfo abci.VoteInfo

// BlockIDFlag indicates which BlockID the signature is for
type BlockIDFlag abci.BlockIDFlag

const (
	BlockIDFlagUnknown = abci.BlockIDFlagUnknown
	// BlockIDFlagAbsent - no vote was received from a validator.
	BlockIDFlagAbsent = abci.BlockIDFlagAbsent
	// BlockIDFlagCommit - voted for the Commit.BlockID.
	BlockIDFlagCommit = abci.BlockIDFlagCommit
	// BlockIDFlagNil - voted for nil.
	BlockIDFlagNil = abci.BlockIDFlagNil
)

// Validator is the validator information of ABCI
type Validator abci.Validator
