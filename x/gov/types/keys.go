package types

import (
	"encoding/binary"
	"fmt"
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
	KeyNextProposalID           = []byte{0x03}

	DepositsKeyPrefix = []byte{0x10}

	VotesKeyPrefix = []byte{0x20}
)

// KeyProposal gets a specific proposal from the store
func KeyProposal(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(ProposalsKeyPrefix, bz...)
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
func KeyInactiveProposalQueue(proposalID uint64, endTime time.Time) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)

	return append(
		append(InactiveProposalQueuePrefix, bz...),
		sdk.FormatTimeBytes(endTime)...,
	)
}

// KeyProposalDeposits gets the first part of the deposits key based on the proposalID
func KeyProposalDeposits(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(DepositsKeyPrefix, bz...)
}

// KeyProposalDeposit key of a specific deposit from the store
func KeyProposalDeposit(proposalID uint64, depositorAddr sdk.AccAddress) []byte {
	return append(KeyProposalDeposits(proposalID), depositorAddr.Bytes()...)
}

// KeyProposalVotes gets the first part of the votes key based on the proposalID
func KeyProposalVotes(proposalID uint64) []byte {
	bz := make([]byte, 8)
	binary.LittleEndian.PutUint64(bz, proposalID)
	return append(VotesKeyPrefix, bz...)
}

// KeyProposalVote key of a specific vote from the store
func KeyProposalVote(proposalID uint64, voterAddr sdk.AccAddress) []byte {
	return append(KeyProposalVotes(proposalID), voterAddr.Bytes()...)
}

// Split keys function; used for iterators

// SplitKeyProposal split the proposal key and returns the proposal id
func SplitKeyProposal(key []byte) (proposalID uint64) {
	if len(key[1:]) != 8 {
		panic(fmt.Sprintf("unexpected key length (%d ≠ 8)", len(key)))
	}

	binary.LittleEndian.PutUint64(key[1:], proposalID)
	return
}

// SplitKeyActiveProposalQueue split the active proposal key and returns the proposal id and endTime
func SplitKeyActiveProposalQueue(key []byte) (proposalID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

// SplitKeyInactiveProposalQueue split the inactive proposal key and returns the proposal id and endTime
func SplitKeyInactiveProposalQueue(key []byte) (proposalID uint64, endTime time.Time) {
	return splitKeyWithTime(key)
}

// SplitKeyDeposit split the deposits key and returns the proposal id and depositor address
func SplitKeyDeposit(key []byte) (proposalID uint64, depositorAddr sdk.AccAddress) {
	return splitKeyWithAddress(key)
}

// SplitKeyVote split the votes key and returns the proposal id and voter address
func SplitKeyVote(key []byte) (proposalID uint64, voterAddr sdk.AccAddress) {
	return splitKeyWithAddress(key)
}

// private functions

func splitKeyWithTime(key []byte) (proposalID uint64, endTime time.Time) {
	lenTime := len(sdk.FormatTimeBytes(time.Now()))
	if len(key[1:]) != 8+lenTime {
		panic(fmt.Sprintf("unexpected key length (%d ≠ %d)", len(key), lenTime+8))
	}

	binary.LittleEndian.PutUint64(key[1:], proposalID)
	endTime, err := sdk.ParseTimeBytes(key[9:])
	if err != nil {
		panic(err)
	}
	return
}

func splitKeyWithAddress(key []byte) (proposalID uint64, addr sdk.AccAddress) {
	if len(key[1:]) != 8+sdk.AddrLen {
		panic(fmt.Sprintf("unexpected key length (%d ≠ 28)", len(key)))
	}
	binary.LittleEndian.PutUint64(key[1:9], proposalID)

	addr = sdk.AccAddress(key[9:])
	return
}
