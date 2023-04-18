package comet

import (
	"context"
	"time"
)

type Service interface {
	GetCometInfo(context.Context) Info
}

type Info struct {
	Evidence []Misbehavior // Evidence misbehavior of the block
	// ValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validator set
	ValidatorsHash    []byte
	ProposerAddress   []byte     // ProposerAddress returns the address of the block proposer
	DecidedLastCommit CommitInfo // DecidedLastCommit returns the last commit info
}

// MisbehaviorType is the type of misbehavior for a validator
type MisbehaviorType int32

const (
	Unknown           MisbehaviorType = 0
	DuplicateVote     MisbehaviorType = 1
	LightClientAttack MisbehaviorType = 2
)

// Validator is the validator information of ABCI
type Validator struct {
	Address []byte
	Power   int64
}

// Misbehavior is the misbehavior information of ABCI
type Misbehavior struct {
	Type             MisbehaviorType
	Validator        Validator
	Height           int64
	Time             time.Time
	TotalVotingPower int64
}

// CommitInfo is the commit information of ABCI
type CommitInfo struct {
	Round int32
	Votes []*VoteInfo
}

// VoteInfo is the vote information of ABCI
type VoteInfo struct {
	Validator       Validator
	SignedLastBlock bool
}
