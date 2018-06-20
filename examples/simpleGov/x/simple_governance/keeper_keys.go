package simpleGovernance

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateProposalKey creates a key of the form "proposals"|{proposalID}
func GenerateProposalKey(proposalID int64) []byte {
	var key []byte
	proposalIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBytes, uint64(proposalID))

	key = []byte("proposals")
	key = append(key, proposalIDBytes...)
	return key
}

// GenerateProposalVotesKey creates a key of the form "proposals"|{proposalID}|"votes"
func GenerateProposalVotesKey(proposalID int64) []byte {
	key := GenerateProposalKey(proposalID)
	key = append(key, []byte("votes")...)
	return key
}

// GenerateProposalVoteKey creates a key of the form "proposals"|{proposalID}|"votes"|{voterAddress}
func GenerateProposalVoteKey(proposalID int64, voterAddr sdk.Address) []byte {
	key := GenerateProposalVotesKey(proposalID)
	key = append(key, voterAddr.Bytes()...)
	return key
}
