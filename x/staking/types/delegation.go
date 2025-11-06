// Package types defines the core data structures for the staking module.
// This file provides delegation types, including standard delegations, unbonding delegations,
// and redelegations that track delegator-validator relationships and token movements.
package types

import (
	"encoding/json"
	"slices"
	"strings"
	"time"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Implements Delegation interface
var _ DelegationI = Delegation{}

// NewDelegation creates a new delegation object with the given delegator address,
// validator address, and initial shares. The delegation represents a delegator's
// stake in a validator and tracks the shares allocated to the delegator.
func NewDelegation(delegatorAddr, validatorAddr string, shares math.LegacyDec) Delegation {
	return Delegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Shares:           shares,
	}
}

// MustMarshalDelegation marshals a delegation to bytes using the provided codec.
// Panics if marshaling fails.
func MustMarshalDelegation(cdc codec.BinaryCodec, delegation Delegation) []byte {
	return cdc.MustMarshal(&delegation)
}

// MustUnmarshalDelegation unmarshals a delegation from bytes using the provided codec.
// Panics if unmarshaling fails.
func MustUnmarshalDelegation(cdc codec.BinaryCodec, value []byte) Delegation {
	delegation, err := UnmarshalDelegation(cdc, value)
	if err != nil {
		panic(err)
	}

	return delegation
}

// UnmarshalDelegation unmarshals a delegation from bytes using the provided codec.
// Returns an error if unmarshaling fails.
func UnmarshalDelegation(cdc codec.BinaryCodec, value []byte) (delegation Delegation, err error) {
	err = cdc.Unmarshal(value, &delegation)
	return delegation, err
}

func (d Delegation) GetDelegatorAddr() string {
	return d.DelegatorAddress
}

func (d Delegation) GetValidatorAddr() string {
	return d.ValidatorAddress
}
func (d Delegation) GetShares() math.LegacyDec { return d.Shares }

// Delegations represents a collection of delegation objects.
// It provides methods for string representation and iteration.
type Delegations []Delegation

func (d Delegations) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// NewUnbondingDelegationEntry creates a new unbonding delegation entry with the given
// creation height, completion time, balance, and unbonding ID. The entry tracks a single
// unbonding operation that will complete at the specified time.
func NewUnbondingDelegationEntry(creationHeight int64, completionTime time.Time, balance math.Int, unbondingID uint64) UnbondingDelegationEntry {
	return UnbondingDelegationEntry{
		CreationHeight:          creationHeight,
		CompletionTime:          completionTime,
		InitialBalance:          balance,
		Balance:                 balance,
		UnbondingId:             unbondingID,
		UnbondingOnHoldRefCount: 0,
	}
}

// IsMature checks if the unbonding delegation entry has reached its completion time.
// An entry is mature when the current time is equal to or after the completion time.
func (e UnbondingDelegationEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// OnHold checks if the unbonding delegation entry is currently on hold.
// An entry is on hold when external modules have placed a hold on it, preventing completion.
func (e UnbondingDelegationEntry) OnHold() bool {
	return e.UnbondingOnHoldRefCount > 0
}

// MustMarshalUBDE marshals an unbonding delegation entry to bytes using the provided codec.
// Panics if marshaling fails.
func MustMarshalUBDE(cdc codec.BinaryCodec, ubd UnbondingDelegationEntry) []byte {
	return cdc.MustMarshal(&ubd)
}

// MustUnmarshalUBDE unmarshals an unbonding delegation entry from bytes using the provided codec.
// Panics if unmarshaling fails.
func MustUnmarshalUBDE(cdc codec.BinaryCodec, value []byte) UnbondingDelegationEntry {
	ubd, err := UnmarshalUBDE(cdc, value)
	if err != nil {
		panic(err)
	}

	return ubd
}

// UnmarshalUBDE unmarshals an unbonding delegation entry from bytes using the provided codec.
// Returns an error if unmarshaling fails.
func UnmarshalUBDE(cdc codec.BinaryCodec, value []byte) (ubd UnbondingDelegationEntry, err error) {
	err = cdc.Unmarshal(value, &ubd)
	return ubd, err
}

// NewUnbondingDelegation creates a new unbonding delegation object with the given
// delegator and validator addresses, creation height, completion time, balance, and unbonding ID.
// The unbonding delegation tracks tokens that are in the process of being unbonded.
func NewUnbondingDelegation(
	delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	creationHeight int64, minTime time.Time, balance math.Int, id uint64,
	valAc, delAc address.Codec,
) UnbondingDelegation {
	valAddr, err := valAc.BytesToString(validatorAddr)
	if err != nil {
		panic(err)
	}
	delAddr, err := delAc.BytesToString(delegatorAddr)
	if err != nil {
		panic(err)
	}
	return UnbondingDelegation{
		DelegatorAddress: delAddr,
		ValidatorAddress: valAddr,
		Entries: []UnbondingDelegationEntry{
			NewUnbondingDelegationEntry(creationHeight, minTime, balance, id),
		},
	}
}

// AddEntry adds a new unbonding entry to the unbonding delegation. If an entry with the same
// creation height and completion time already exists, it merges the balances. Returns true
// if a new entry was created, false if an existing entry was updated.
func (ubd *UnbondingDelegation) AddEntry(creationHeight int64, minTime time.Time, balance math.Int, unbondingID uint64) bool {
	// Check the entries exists with creation_height and complete_time
	entryIndex := -1
	for index, ubdEntry := range ubd.Entries {
		if ubdEntry.CreationHeight == creationHeight && ubdEntry.CompletionTime.Equal(minTime) {
			entryIndex = index
			break
		}
	}
	// entryIndex exists
	if entryIndex != -1 {
		ubdEntry := ubd.Entries[entryIndex]
		ubdEntry.Balance = ubdEntry.Balance.Add(balance)
		ubdEntry.InitialBalance = ubdEntry.InitialBalance.Add(balance)

		// update the entry
		ubd.Entries[entryIndex] = ubdEntry
		return false
	}
	// append the new unbond delegation entry
	entry := NewUnbondingDelegationEntry(creationHeight, minTime, balance, unbondingID)
	ubd.Entries = append(ubd.Entries, entry)
	return true
}

// RemoveEntry removes the unbonding delegation entry at the specified index.
// The entry is deleted from the entries slice.
func (ubd *UnbondingDelegation) RemoveEntry(i int64) {
	ubd.Entries = slices.Delete(ubd.Entries, int(i), int(i+1))
}

// MustMarshalUBD marshals an unbonding delegation to bytes using the provided codec.
// Panics if marshaling fails.
func MustMarshalUBD(cdc codec.BinaryCodec, ubd UnbondingDelegation) []byte {
	return cdc.MustMarshal(&ubd)
}

// MustUnmarshalUBD unmarshals an unbonding delegation from bytes using the provided codec.
// Panics if unmarshaling fails.
func MustUnmarshalUBD(cdc codec.BinaryCodec, value []byte) UnbondingDelegation {
	ubd, err := UnmarshalUBD(cdc, value)
	if err != nil {
		panic(err)
	}

	return ubd
}

// UnmarshalUBD unmarshals an unbonding delegation from bytes using the provided codec.
// Returns an error if unmarshaling fails.
func UnmarshalUBD(cdc codec.BinaryCodec, value []byte) (ubd UnbondingDelegation, err error) {
	err = cdc.Unmarshal(value, &ubd)
	return ubd, err
}

// UnbondingDelegations represents a collection of unbonding delegation objects.
// It provides methods for string representation and iteration.
type UnbondingDelegations []UnbondingDelegation

func (ubds UnbondingDelegations) String() (out string) {
	for _, u := range ubds {
		out += u.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// NewRedelegationEntry creates a new redelegation entry with the given creation height,
// completion time, balance, destination shares, and unbonding ID. The entry tracks a single
// redelegation operation that will complete at the specified time.
func NewRedelegationEntry(creationHeight int64, completionTime time.Time, balance math.Int, sharesDst math.LegacyDec, id uint64) RedelegationEntry {
	return RedelegationEntry{
		CreationHeight:          creationHeight,
		CompletionTime:          completionTime,
		InitialBalance:          balance,
		SharesDst:               sharesDst,
		UnbondingId:             id,
		UnbondingOnHoldRefCount: 0,
	}
}

// IsMature checks if the redelegation entry has reached its completion time.
// An entry is mature when the current time is equal to or after the completion time.
func (e RedelegationEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// OnHold checks if the redelegation entry is currently on hold.
// An entry is on hold when external modules have placed a hold on it, preventing completion.
func (e RedelegationEntry) OnHold() bool {
	return e.UnbondingOnHoldRefCount > 0
}

// NewRedelegation creates a new redelegation object with the given delegator address,
// source and destination validator addresses, creation height, completion time, balance,
// destination shares, and unbonding ID. The redelegation tracks tokens being moved between validators.
func NewRedelegation(
	delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress,
	creationHeight int64, minTime time.Time, balance math.Int, sharesDst math.LegacyDec, id uint64,
	valAc, delAc address.Codec,
) Redelegation {
	valSrcAddr, err := valAc.BytesToString(validatorSrcAddr)
	if err != nil {
		panic(err)
	}
	valDstAddr, err := valAc.BytesToString(validatorDstAddr)
	if err != nil {
		panic(err)
	}
	delAddr, err := delAc.BytesToString(delegatorAddr)
	if err != nil {
		panic(err)
	}

	return Redelegation{
		DelegatorAddress:    delAddr,
		ValidatorSrcAddress: valSrcAddr,
		ValidatorDstAddress: valDstAddr,
		Entries: []RedelegationEntry{
			NewRedelegationEntry(creationHeight, minTime, balance, sharesDst, id),
		},
	}
}

// AddEntry adds a new redelegation entry to the redelegation object.
// The entry is appended to the entries slice and tracks a single redelegation operation.
func (red *Redelegation) AddEntry(creationHeight int64, minTime time.Time, balance math.Int, sharesDst math.LegacyDec, id uint64) {
	entry := NewRedelegationEntry(creationHeight, minTime, balance, sharesDst, id)
	red.Entries = append(red.Entries, entry)
}

// RemoveEntry removes the redelegation entry at the specified index.
// The entry is deleted from the entries slice.
func (red *Redelegation) RemoveEntry(i int64) {
	red.Entries = slices.Delete(red.Entries, int(i), int(i+1))
}

// MustMarshalRED marshals a redelegation to bytes using the provided codec.
// Panics if marshaling fails.
func MustMarshalRED(cdc codec.BinaryCodec, red Redelegation) []byte {
	return cdc.MustMarshal(&red)
}

// MustUnmarshalRED unmarshals a redelegation from bytes using the provided codec.
// Panics if unmarshaling fails.
func MustUnmarshalRED(cdc codec.BinaryCodec, value []byte) Redelegation {
	red, err := UnmarshalRED(cdc, value)
	if err != nil {
		panic(err)
	}

	return red
}

// UnmarshalRED unmarshals a redelegation from bytes using the provided codec.
// Returns an error if unmarshaling fails.
func UnmarshalRED(cdc codec.BinaryCodec, value []byte) (red Redelegation, err error) {
	err = cdc.Unmarshal(value, &red)
	return red, err
}

// Redelegations represents a collection of redelegation objects.
// It provides methods for string representation and iteration.
type Redelegations []Redelegation

func (d Redelegations) String() (out string) {
	for _, red := range d {
		out += red.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// ----------------------------------------------------------------------------
// Client Types

// NewDelegationResp creates a new DelegationResponse instance with the given delegator address,
// validator address, shares, and balance. The response combines delegation information with
// the current token balance for client queries.
func NewDelegationResp(
	delegatorAddr, validatorAddr string, shares math.LegacyDec, balance sdk.Coin,
) DelegationResponse {
	return DelegationResponse{
		Delegation: NewDelegation(delegatorAddr, validatorAddr, shares),
		Balance:    balance,
	}
}

type delegationRespAlias DelegationResponse

// MarshalJSON implements the json.Marshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (d DelegationResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((delegationRespAlias)(d))
}

// UnmarshalJSON implements the json.Unmarshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (d *DelegationResponse) UnmarshalJSON(bz []byte) error {
	return json.Unmarshal(bz, (*delegationRespAlias)(d))
}

// DelegationResponses represents a collection of delegation response objects.
// It provides methods for string representation and iteration.
type DelegationResponses []DelegationResponse

// String implements the Stringer interface for DelegationResponses.
func (d DelegationResponses) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// NewRedelegationResponse creates a new RedelegationResponse instance with the given delegator address,
// source and destination validator addresses, and redelegation entry responses. The response combines
// redelegation information with entry details for client queries.
func NewRedelegationResponse(
	delegatorAddr, validatorSrc, validatorDst string, entries []RedelegationEntryResponse,
) RedelegationResponse {
	return RedelegationResponse{
		Redelegation: Redelegation{
			DelegatorAddress:    delegatorAddr,
			ValidatorSrcAddress: validatorSrc,
			ValidatorDstAddress: validatorDst,
		},
		Entries: entries,
	}
}

// NewRedelegationEntryResponse creates a new RedelegationEntryResponse instance with the given
// creation height, completion time, destination shares, initial balance, current balance, and
// unbonding ID. The response combines redelegation entry information with the current balance.
func NewRedelegationEntryResponse(
	creationHeight int64, completionTime time.Time, sharesDst math.LegacyDec, initialBalance, balance math.Int, unbondingID uint64,
) RedelegationEntryResponse {
	return RedelegationEntryResponse{
		RedelegationEntry: NewRedelegationEntry(creationHeight, completionTime, initialBalance, sharesDst, unbondingID),
		Balance:           balance,
	}
}

type redelegationRespAlias RedelegationResponse

// MarshalJSON implements the json.Marshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (r RedelegationResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal((redelegationRespAlias)(r))
}

// UnmarshalJSON implements the json.Unmarshaler interface. This is so we can
// achieve a flattened structure while embedding other types.
func (r *RedelegationResponse) UnmarshalJSON(bz []byte) error {
	return json.Unmarshal(bz, (*redelegationRespAlias)(r))
}

// RedelegationResponses represents a collection of redelegation response objects.
// It provides methods for string representation and iteration.
type RedelegationResponses []RedelegationResponse

func (r RedelegationResponses) String() (out string) {
	for _, red := range r {
		out += red.String() + "\n"
	}

	return strings.TrimSpace(out)
}
