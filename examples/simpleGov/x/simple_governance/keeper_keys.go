package simpleGovernance

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GenerateProposalKey(proposalID int64) []byte {
	var key []byte
	proposalIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(proposalIDBytes, uint64(proposalID))

	// key is of the form "proposals"|{proposalID}
	key = []byte("proposals")
	key = append(key, proposalIDBytes...)
	return key
}

func GenerateAccountProposalsKey(voterAddr sdk.Address) []byte {
	// key is of the form "accounts"|{voterAddress}|"proposals"
	key := []byte("accounts")
	key = append(key, voterAddr.Bytes()...)
	key = append(key, []byte("proposals")...)
	return key
}

func GenerateAccountProposalKey(proposalID int64, voterAddr sdk.Address) []byte {
	// key is of the form "accounts"|{voterAddress}|"proposals"|{proposalID}
	key := GenerateAccountProposalsKey(voterAddr)
	key = append(key, GenerateProposalKey(proposalID)...)
	return key
}

func GenerateAccountProposalsVoteKey(proposalID int64, voterAddr sdk.Address) []byte {
	// key is of the form "accounts"|{voterAddress}|"proposals"|{proposalID}|"vote"
	key := GenerateAccountProposalKey(proposalID, voterAddr)
	key = append(key, []byte("vote")...)
	return key
}
