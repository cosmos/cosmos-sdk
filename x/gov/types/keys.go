package types

import (
	"encoding/binary"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// DefaultParamspace default name for parameter store
	DefaultParamspace = ModuleName
)

// Keys for governance store
var (
	ProposalsKeyPrefix          = []byte{0x00}
	ActiveProposalQueuePrefix   = []byte{0x01}
	InactiveProposalQueuePrefix = []byte{0x02}

	DepositsKeyPrefix = []byte{0x10}

	VotesKeyPrefix = []byte{0x20}
)

// KeyProposal gets a specific proposal from the store
func KeyProposal(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(ProposalsKeyPrefix, bz...)
}

// KeyDeposit gets a specific deposit from the store
func KeyDeposit(proposalID uint64, depositorAddr sdk.AccAddress) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(append(DepositsKeyPrefix, bz...), depositorAddr.Bytes()...)
}

// KeyVote gets  a specific vote from the store
func KeyVote(proposalID uint64, voterAddr sdk.AccAddress) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(append(VotesKeyPrefix, bz...), voterAddr.Bytes()...)
}

// KeyActiveProposalQueue returns the key for a proposalID in the activeProposalQueue
func KeyActiveProposalQueue(proposalID uint64, endTime time.Time) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)

	return append(
		append(ActiveProposalQueuePrefix, bz...),
		sdk.FormatTimeBytes(endTime)...,
	)
}

// KeyInactiveProposalQueueProposal returns the key for a proposalID in the inactiveProposalQueue
func KeyInactiveProposalQueueProposal(proposalID uint64, endTime time.Time) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)

	return append(
		append(InactiveProposalQueuePrefix, bz...),
		sdk.FormatTimeBytes(endTime)...,
	)

}
