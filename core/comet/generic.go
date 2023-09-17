package comet

import (
	"cosmossdk.io/core/header"
)

type CometHeader struct {
	// Height  uint64    // GetHeight returns the height of the block
	// Hash    []byte    // GetHash returns the hash of the block header
	// Time    time.Time // GetTime returns the time of the block
	// ChainID string    // GetChainID returns the chain ID of the chain
	// AppHash []byte    // GetAppHash used in the current block header
	// embed native header
	header.Info

	// Specifc to Comet info
	Evidence []Evidence // Evidence misbehavior of the block
	// ValidatorsHash returns the hash of the validators
	// For Comet, it is the hash of the next validator set
	ValidatorsHash  []byte
	ProposerAddress []byte     // ProposerAddress is  the address of the block proposer
	LastCommit      CommitInfo // DecidedLastCommit returns the last commit info
}
