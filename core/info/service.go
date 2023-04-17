package info

import (
	"context"
	"time"
)

type BlockService interface {
	GetBlockInfo(context.Context) BlockInfo
}

type BlockInfo interface {
	GetHeight() int64      // GetHeight returns the height of the block
	GetHeaderHash() []byte // GetHeaderHash returns the hash of the block header
	GetTime() time.Time    // GetTime returns the time of the block
	GetChainID() string    // GetChainId returns the chain ID of the block
}

type CometService interface {
	GetCometInfo(context.Context) CometInfo
}

type CometInfo interface {
	GetMisbehavior() []Misbehavior // Misbehavior returns the misbehavior of the block
	// GetValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validator set
	GetValidatorsHash() []byte
	GetProposerAddress() []byte       // GetProposerAddress returns the address of the block proposer
	GetDecidedLastCommit() CommitInfo // GetDecidedLastCommit returns the last commit info
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
