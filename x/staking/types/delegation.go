package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// DVPair is struct that just has a delegator-validator pair with no other data.
// It is intended to be used as a marshalable pointer. For example, a DVPair can be used to construct the
// key to getting an UnbondingDelegation from state.
type DVPair struct {
	DelegatorAddress sdk.AccAddress
	ValidatorAddress sdk.ValAddress
}

// DVVTriplet is struct that just has a delegator-validator-validator triplet with no other data.
// It is intended to be used as a marshalable pointer. For example, a DVVTriplet can be used to construct the
// key to getting a Redelegation from state.
type DVVTriplet struct {
	DelegatorAddress    sdk.AccAddress
	ValidatorSrcAddress sdk.ValAddress
	ValidatorDstAddress sdk.ValAddress
}

// Implements Delegation interface
var _ exported.DelegationI = Delegation{}

// Delegation represents the bond with tokens held by an account. It is
// owned by one delegator, and is associated with the voting power of one
// validator.
type Delegation struct {
	DelegatorAddress sdk.AccAddress `json:"delegator_address"`
	ValidatorAddress sdk.ValAddress `json:"validator_address"`
	Shares           sdk.Dec        `json:"shares"`
}

// NewDelegation creates a new delegation object
func NewDelegation(delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress,
	shares sdk.Dec) Delegation {

	return Delegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Shares:           shares,
	}
}

// return the delegation
func MustMarshalDelegation(cdc *codec.Codec, delegation Delegation) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(delegation)
}

// return the delegation
func MustUnmarshalDelegation(cdc *codec.Codec, value []byte) Delegation {
	delegation, err := UnmarshalDelegation(cdc, value)
	if err != nil {
		panic(err)
	}
	return delegation
}

// return the delegation
func UnmarshalDelegation(cdc *codec.Codec, value []byte) (delegation Delegation, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &delegation)
	return delegation, err
}

// nolint
func (d Delegation) Equal(d2 Delegation) bool {
	return bytes.Equal(d.DelegatorAddress, d2.DelegatorAddress) &&
		bytes.Equal(d.ValidatorAddress, d2.ValidatorAddress) &&
		d.Shares.Equal(d2.Shares)
}

// nolint - for Delegation
func (d Delegation) GetDelegatorAddr() sdk.AccAddress { return d.DelegatorAddress }
func (d Delegation) GetValidatorAddr() sdk.ValAddress { return d.ValidatorAddress }
func (d Delegation) GetShares() sdk.Dec               { return d.Shares }

// String returns a human readable string representation of a Delegation.
func (d Delegation) String() string {
	return fmt.Sprintf(`Delegation:
  Delegator: %s
  Validator: %s
  Shares:    %s`, d.DelegatorAddress,
		d.ValidatorAddress, d.Shares)
}

// Delegations is a collection of delegations
type Delegations []Delegation

func (d Delegations) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// UnbondingDelegation stores all of a single delegator's unbonding bonds
// for a single validator in an time-ordered list
type UnbondingDelegation struct {
	DelegatorAddress sdk.AccAddress             `json:"delegator_address"` // delegator
	ValidatorAddress sdk.ValAddress             `json:"validator_address"` // validator unbonding from operator addr
	Entries          []UnbondingDelegationEntry `json:"entries"`           // unbonding delegation entries
}

// UnbondingDelegationEntry - entry to an UnbondingDelegation
type UnbondingDelegationEntry struct {
	CreationHeight int64     `json:"creation_height"` // height which the unbonding took place
	CompletionTime time.Time `json:"completion_time"` // time at which the unbonding delegation will complete
	InitialBalance sdk.Int   `json:"initial_balance"` // atoms initially scheduled to receive at completion
	Balance        sdk.Int   `json:"balance"`         // atoms to receive at completion
}

// IsMature - is the current entry mature
func (e UnbondingDelegationEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// NewUnbondingDelegation - create a new unbonding delegation object
func NewUnbondingDelegation(delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress, creationHeight int64, minTime time.Time,
	balance sdk.Int) UnbondingDelegation {

	entry := NewUnbondingDelegationEntry(creationHeight, minTime, balance)
	return UnbondingDelegation{
		DelegatorAddress: delegatorAddr,
		ValidatorAddress: validatorAddr,
		Entries:          []UnbondingDelegationEntry{entry},
	}
}

// NewUnbondingDelegation - create a new unbonding delegation object
func NewUnbondingDelegationEntry(creationHeight int64, completionTime time.Time,
	balance sdk.Int) UnbondingDelegationEntry {

	return UnbondingDelegationEntry{
		CreationHeight: creationHeight,
		CompletionTime: completionTime,
		InitialBalance: balance,
		Balance:        balance,
	}
}

// AddEntry - append entry to the unbonding delegation
func (d *UnbondingDelegation) AddEntry(creationHeight int64,
	minTime time.Time, balance sdk.Int) {

	entry := NewUnbondingDelegationEntry(creationHeight, minTime, balance)
	d.Entries = append(d.Entries, entry)
}

// RemoveEntry - remove entry at index i to the unbonding delegation
func (d *UnbondingDelegation) RemoveEntry(i int64) {
	d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
}

// return the unbonding delegation
func MustMarshalUBD(cdc *codec.Codec, ubd UnbondingDelegation) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(ubd)
}

// unmarshal a unbonding delegation from a store value
func MustUnmarshalUBD(cdc *codec.Codec, value []byte) UnbondingDelegation {
	ubd, err := UnmarshalUBD(cdc, value)
	if err != nil {
		panic(err)
	}
	return ubd
}

// unmarshal a unbonding delegation from a store value
func UnmarshalUBD(cdc *codec.Codec, value []byte) (ubd UnbondingDelegation, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &ubd)
	return ubd, err
}

// nolint
// inefficient but only used in testing
func (d UnbondingDelegation) Equal(d2 UnbondingDelegation) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d2)
	return bytes.Equal(bz1, bz2)
}

// String returns a human readable string representation of an UnbondingDelegation.
func (d UnbondingDelegation) String() string {
	out := fmt.Sprintf(`Unbonding Delegations between:
  Delegator:                 %s
  Validator:                 %s
	Entries:`, d.DelegatorAddress, d.ValidatorAddress)
	for i, entry := range d.Entries {
		out += fmt.Sprintf(`    Unbonding Delegation %d:
      Creation Height:           %v
      Min time to unbond (unix): %v
      Expected balance:          %s`, i, entry.CreationHeight,
			entry.CompletionTime, entry.Balance)
	}
	return out
}

// UnbondingDelegations is a collection of UnbondingDelegation
type UnbondingDelegations []UnbondingDelegation

func (ubds UnbondingDelegations) String() (out string) {
	for _, u := range ubds {
		out += u.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// Redelegation contains the list of a particular delegator's
// redelegating bonds from a particular source validator to a
// particular destination validator
type Redelegation struct {
	DelegatorAddress    sdk.AccAddress      `json:"delegator_address"`     // delegator
	ValidatorSrcAddress sdk.ValAddress      `json:"validator_src_address"` // validator redelegation source operator addr
	ValidatorDstAddress sdk.ValAddress      `json:"validator_dst_address"` // validator redelegation destination operator addr
	Entries             []RedelegationEntry `json:"entries"`               // redelegation entries
}

// RedelegationEntry - entry to a Redelegation
type RedelegationEntry struct {
	CreationHeight int64     `json:"creation_height"` // height at which the redelegation took place
	CompletionTime time.Time `json:"completion_time"` // time at which the redelegation will complete
	InitialBalance sdk.Int   `json:"initial_balance"` // initial balance when redelegation started
	SharesDst      sdk.Dec   `json:"shares_dst"`      // amount of destination-validator shares created by redelegation
}

// NewRedelegation - create a new redelegation object
func NewRedelegation(delegatorAddr sdk.AccAddress, validatorSrcAddr,
	validatorDstAddr sdk.ValAddress, creationHeight int64,
	minTime time.Time, balance sdk.Int,
	sharesDst sdk.Dec) Redelegation {

	entry := NewRedelegationEntry(creationHeight,
		minTime, balance, sharesDst)

	return Redelegation{
		DelegatorAddress:    delegatorAddr,
		ValidatorSrcAddress: validatorSrcAddr,
		ValidatorDstAddress: validatorDstAddr,
		Entries:             []RedelegationEntry{entry},
	}
}

// NewRedelegation - create a new redelegation object
func NewRedelegationEntry(creationHeight int64,
	completionTime time.Time, balance sdk.Int,
	sharesDst sdk.Dec) RedelegationEntry {

	return RedelegationEntry{
		CreationHeight: creationHeight,
		CompletionTime: completionTime,
		InitialBalance: balance,
		SharesDst:      sharesDst,
	}
}

// IsMature - is the current entry mature
func (e RedelegationEntry) IsMature(currentTime time.Time) bool {
	return !e.CompletionTime.After(currentTime)
}

// AddEntry - append entry to the unbonding delegation
func (d *Redelegation) AddEntry(creationHeight int64,
	minTime time.Time, balance sdk.Int,
	sharesDst sdk.Dec) {

	entry := NewRedelegationEntry(creationHeight, minTime, balance, sharesDst)
	d.Entries = append(d.Entries, entry)
}

// RemoveEntry - remove entry at index i to the unbonding delegation
func (d *Redelegation) RemoveEntry(i int64) {
	d.Entries = append(d.Entries[:i], d.Entries[i+1:]...)
}

// return the redelegation
func MustMarshalRED(cdc *codec.Codec, red Redelegation) []byte {
	return cdc.MustMarshalBinaryLengthPrefixed(red)
}

// unmarshal a redelegation from a store value
func MustUnmarshalRED(cdc *codec.Codec, value []byte) Redelegation {
	red, err := UnmarshalRED(cdc, value)
	if err != nil {
		panic(err)
	}
	return red
}

// unmarshal a redelegation from a store value
func UnmarshalRED(cdc *codec.Codec, value []byte) (red Redelegation, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &red)
	return red, err
}

// nolint
// inefficient but only used in tests
func (d Redelegation) Equal(d2 Redelegation) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&d2)
	return bytes.Equal(bz1, bz2)
}

// String returns a human readable string representation of a Redelegation.
func (d Redelegation) String() string {
	out := fmt.Sprintf(`Redelegations between:
  Delegator:                 %s
  Source Validator:          %s
  Destination Validator:     %s
  Entries:
`,
		d.DelegatorAddress, d.ValidatorSrcAddress, d.ValidatorDstAddress,
	)

	for i, entry := range d.Entries {
		out += fmt.Sprintf(`    Redelegation Entry #%d:
      Creation height:           %v
      Min time to unbond (unix): %v
      Dest Shares:               %s
`,
			i, entry.CreationHeight, entry.CompletionTime, entry.SharesDst,
		)
	}

	return strings.TrimRight(out, "\n")
}

// Redelegations are a collection of Redelegation
type Redelegations []Redelegation

func (d Redelegations) String() (out string) {
	for _, red := range d {
		out += red.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// ----------------------------------------------------------------------------
// Client Types

// DelegationResponse is equivalent to Delegation except that it contains a balance
// in addition to shares which is more suitable for client responses.
type DelegationResponse struct {
	Delegation
	Balance sdk.Int `json:"balance"`
}

func NewDelegationResp(d sdk.AccAddress, v sdk.ValAddress, s sdk.Dec, b sdk.Int) DelegationResponse {
	return DelegationResponse{NewDelegation(d, v, s), b}
}

// String implements the Stringer interface for DelegationResponse.
func (d DelegationResponse) String() string {
	return fmt.Sprintf("%s\n  Balance:   %s", d.Delegation.String(), d.Balance)
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

// DelegationResponses is a collection of DelegationResp
type DelegationResponses []DelegationResponse

// String implements the Stringer interface for DelegationResponses.
func (d DelegationResponses) String() (out string) {
	for _, del := range d {
		out += del.String() + "\n"
	}
	return strings.TrimSpace(out)
}

// RedelegationResponse is equivalent to a Redelegation except that its entries
// contain a balance in addition to shares which is more suitable for client
// responses.
type RedelegationResponse struct {
	Redelegation
	Entries []RedelegationEntryResponse `json:"entries"` // nolint: structtag
}

// RedelegationEntryResponse is equivalent to a RedelegationEntry except that it
// contains a balance in addition to shares which is more suitable for client
// responses.
type RedelegationEntryResponse struct {
	RedelegationEntry
	Balance sdk.Int `json:"balance"`
}

func NewRedelegationResponse(d sdk.AccAddress, vSrc, vDst sdk.ValAddress, entries []RedelegationEntryResponse) RedelegationResponse {
	return RedelegationResponse{
		Redelegation{
			DelegatorAddress:    d,
			ValidatorSrcAddress: vSrc,
			ValidatorDstAddress: vDst,
		},
		entries,
	}
}

func NewRedelegationEntryResponse(ch int64, ct time.Time, s sdk.Dec, ib, b sdk.Int) RedelegationEntryResponse {
	return RedelegationEntryResponse{NewRedelegationEntry(ch, ct, ib, s), b}
}

// String implements the Stringer interface for RedelegationResp.
func (r RedelegationResponse) String() string {
	out := fmt.Sprintf(`Redelegations between:
  Delegator:                 %s
  Source Validator:          %s
  Destination Validator:     %s
  Entries:
`,
		r.DelegatorAddress, r.ValidatorSrcAddress, r.ValidatorDstAddress,
	)

	for i, entry := range r.Entries {
		out += fmt.Sprintf(`    Redelegation Entry #%d:
      Creation height:           %v
      Min time to unbond (unix): %v
      Initial Balance:           %s
      Shares:                    %s
      Balance:                   %s
`,
			i, entry.CreationHeight, entry.CompletionTime, entry.InitialBalance, entry.SharesDst, entry.Balance,
		)
	}

	return strings.TrimRight(out, "\n")
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

// RedelegationResponses are a collection of RedelegationResp
type RedelegationResponses []RedelegationResponse

func (r RedelegationResponses) String() (out string) {
	for _, red := range r {
		out += red.String() + "\n"
	}
	return strings.TrimSpace(out)
}
