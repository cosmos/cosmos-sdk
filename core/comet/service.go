package comet

import (
	"context"
	"time"
)

// Service is an interface that can be used to get information specific to Comet
type InfoService interface {
	GetCometInfo(context.Context) BlockInfo
}

// BlockInfo is the information comet provides apps in ABCI
type BlockInfo interface {
	GetEvidence() []Misbehavior // Evidence misbehavior of the block
	// ValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validator set
	GetValidatorsHash() []byte
	GetProposerAddress() []byte // ProposerAddress returns the address of the block proposer
	GetLastCommit() CommitInfo  // DecidedLastCommit returns the last commit info
}

// MisbehaviorType is the type of misbehavior for a validator
type MisbehaviorType int32

const (
	Unknown           MisbehaviorType = 0
	DuplicateVote     MisbehaviorType = 1
	LightClientAttack MisbehaviorType = 2
)

// Validator is the validator information of ABCI
type Validator interface {
	Address() []byte
	Power() int64
}

// Misbehavior is the misbehavior information of ABCI
type Misbehavior interface {
	Type() MisbehaviorType
	Validator() Validator
	Height() int64
	Time() time.Time
	TotalVotingPower() int64
}

// CommitInfo is the commit information of ABCI
type CommitInfo interface {
	Round() int32
	Votes() VoteInfos
}

// VoteInfos is an interface to get specific votes in a efficient way
type VoteInfos interface {
	Len() int
	Get(int) VoteInfo
}

// VoteInfo is the vote information of ABCI
type VoteInfo interface {
	Validator() Validator
	SignedLastBlock() bool
}
