package types

import (
	"encoding/binary"
	"time"

	"cosmossdk.io/collections"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
)

const (
	// ModuleName is the name of the module
	ModuleName = "gov"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName
)

// Keys for governance store
// Items are stored with the following key: values
//
// - 0x00<proposalID_Bytes>: Proposal
//
// - 0x01<endTime_Bytes><proposalID_Bytes>: activeProposalID
//
// - 0x02<endTime_Bytes><proposalID_Bytes>: inactiveProposalID
//
// - 0x03: nextProposalID
//
// - 0x04<proposalID_Bytes>: []byte{0x01} if proposalID is in the voting period
//
// - 0x10<proposalID_Bytes><depositorAddrLen (1 Byte)><depositorAddr_Bytes>: Deposit
//
// - 0x20<proposalID_Bytes><voterAddrLen (1 Byte)><voterAddr_Bytes>: Voter
//
// - 0x30: Params
var (
	ProposalsKeyPrefix            = collections.NewPrefix(0)
	ActiveProposalQueuePrefix     = []byte{0x01}
	InactiveProposalQueuePrefix   = []byte{0x02}
	ProposalIDKey                 = collections.NewPrefix(3)
	VotingPeriodProposalKeyPrefix = []byte{0x04}

	DepositsKeyPrefix = collections.NewPrefix(16)

	VotesKeyPrefix = collections.NewPrefix(32)

	// ParamsKey is the key to query all gov params
	ParamsKey = collections.NewPrefix(48)

	// ConstitutionKey is the key string used to store the chain's constitution
	ConstitutionKey = collections.NewPrefix(49)
)

var lenTime = len(sdk.FormatTimeBytes(time.Now()))

// GetProposalIDBytes returns the byte representation of the proposalID
func GetProposalIDBytes(proposalID uint64) (proposalIDBz []byte) {
	proposalIDBz = make([]byte, 8)
	binary.BigEndian.PutUint64(proposalIDBz, proposalID)
	return
}

// GetProposalIDFromBytes returns proposalID in uint64 format from a byte array
func GetProposalIDFromBytes(bz []byte) (proposalID uint64) {
	return binary.BigEndian.Uint64(bz)
}

// ProposalKey gets a specific proposal from the store
func ProposalKey(proposalID uint64) []byte {
	return append(ProposalsKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// VotingPeriodProposalKey gets if a proposal is in voting period.
func VotingPeriodProposalKey(proposalID uint64) []byte {
	return append(VotingPeriodProposalKeyPrefix, GetProposalIDBytes(proposalID)...)
}

// ActiveProposalByTimeKey gets the active proposal queue key by endTime
func ActiveProposalByTimeKey(endTime time.Time) []byte {
	return append(ActiveProposalQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

// ActiveProposalQueueKey returns the key for a proposalID in the activeProposalQueue
func ActiveProposalQueueKey(proposalID uint64, endTime time.Time) []byte {
	return append(ActiveProposalByTimeKey(endTime), GetProposalIDBytes(proposalID)...)
}

// InactiveProposalByTimeKey gets the inactive proposal queue key by endTime
func InactiveProposalByTimeKey(endTime time.Time) []byte {
	return append(InactiveProposalQueuePrefix, sdk.FormatTimeBytes(endTime)...)
}

// InactiveProposalQueueKey returns the key for a proposalID in the inactiveProposalQueue
func InactiveProposalQueueKey(proposalID uint64, endTime time.Time) []byte {
	return append(InactiveProposalByTimeKey(endTime), GetProposalIDBytes(proposalID)...)
}

// Split keys function; used for iterators

// SplitProposalKey split the proposal key and returns the proposal id
func SplitProposalKey(key []byte) (proposalID uint64) {
	kv.AssertKeyLength(key[1:], 8)
	return GetProposalIDFromBytes(key[1:])
}

// SplitActiveProposalQueueKey split the active proposal key and returns the proposal id and endTime
func SplitActiveProposalQueueKey(key []byte) (proposalID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

// SplitInactiveProposalQueueKey split the inactive proposal key and returns the proposal id and endTime
func SplitInactiveProposalQueueKey(key []byte) (proposalID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

// private functions

func splitKeyWithTime(key []byte) (proposalID uint64, endTime time.Time) {
	kv.AssertKeyLength(key[1:], 8+lenTime)

	endTime, err := sdk.ParseTimeBytes(key[1 : 1+lenTime])
	if err != nil {
		panic(err)
	}

	proposalID = GetProposalIDFromBytes(key[1+lenTime:])
	return
}
