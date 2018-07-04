package types

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Delegation represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type Delegation struct {
	DelegatorAddr sdk.Address `json:"delegator_addr"`
	ValidatorAddr sdk.Address `json:"validator_addr"`
	Shares        sdk.Rat     `json:"shares"`
	Height        int64       `json:"height"` // Last height bond updated
}

// Equal returns a boolean determining if two Delegation types are identical.
func (d Delegation) Equal(d2 Delegation) bool {
	return bytes.Equal(d.DelegatorAddr, d2.DelegatorAddr) &&
		bytes.Equal(d.ValidatorAddr, d2.ValidatorAddr) &&
		d.Height == d2.Height &&
		d.Shares.Equal(d2.Shares)
}

// ensure fulfills the sdk validator types
var _ sdk.Delegation = Delegation{}

// nolint - for sdk.Delegation
func (d Delegation) GetDelegator() sdk.Address { return d.DelegatorAddr }
func (d Delegation) GetValidator() sdk.Address { return d.ValidatorAddr }
func (d Delegation) GetBondShares() sdk.Rat    { return d.Shares }

// HumanReadableString returns a human readable string representation of a
// Delegation. An error is returned if the Delegation's delegator or validator
// addresses cannot be Bech32 encoded.
func (d Delegation) HumanReadableString() (string, error) {
	bechAcc, err := sdk.Bech32ifyAcc(d.DelegatorAddr)
	if err != nil {
		return "", err
	}

	bechVal, err := sdk.Bech32ifyAcc(d.ValidatorAddr)
	if err != nil {
		return "", err
	}

	resp := "Delegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", bechAcc)
	resp += fmt.Sprintf("Validator: %s\n", bechVal)
	resp += fmt.Sprintf("Shares: %s", d.Shares.String())
	resp += fmt.Sprintf("Height: %d", d.Height)

	return resp, nil
}

// UnbondingDelegation reflects a delegation's passive unbonding queue.
type UnbondingDelegation struct {
	DelegatorAddr  sdk.Address `json:"delegator_addr"`  // delegator
	ValidatorAddr  sdk.Address `json:"validator_addr"`  // validator unbonding from owner addr
	CreationHeight int64       `json:"creation_height"` // height which the unbonding took place
	MinTime        int64       `json:"min_time"`        // unix time for unbonding completion
	InitialBalance sdk.Coin    `json:"initial_balance"` // atoms initially scheduled to receive at completion
	Balance        sdk.Coin    `json:"balance"`         // atoms to receive at completion
}

// Equal returns a boolean determining if two UnbondingDelegation types are
// identical.
func (d UnbondingDelegation) Equal(d2 UnbondingDelegation) bool {
	bz1 := MsgCdc.MustMarshalBinary(&d)
	bz2 := MsgCdc.MustMarshalBinary(&d2)
	return bytes.Equal(bz1, bz2)
}

// HumanReadableString returns a human readable string representation of an
// UnbondingDelegation. An error is returned if the UnbondingDelegation's
// delegator or validator addresses cannot be Bech32 encoded.
func (d UnbondingDelegation) HumanReadableString() (string, error) {
	bechAcc, err := sdk.Bech32ifyAcc(d.DelegatorAddr)
	if err != nil {
		return "", err
	}

	bechVal, err := sdk.Bech32ifyAcc(d.ValidatorAddr)
	if err != nil {
		return "", err
	}

	resp := "Unbonding Delegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", bechAcc)
	resp += fmt.Sprintf("Validator: %s\n", bechVal)
	resp += fmt.Sprintf("Creation height: %v\n", d.CreationHeight)
	resp += fmt.Sprintf("Min time to unbond (unix): %v\n", d.MinTime)
	resp += fmt.Sprintf("Expected balance: %s", d.Balance.String())

	return resp, nil

}

// Redelegation reflects a delegation's passive re-delegation queue.
type Redelegation struct {
	DelegatorAddr    sdk.Address `json:"delegator_addr"`     // delegator
	ValidatorSrcAddr sdk.Address `json:"validator_src_addr"` // validator redelegation source owner addr
	ValidatorDstAddr sdk.Address `json:"validator_dst_addr"` // validator redelegation destination owner addr
	CreationHeight   int64       `json:"creation_height"`    // height which the redelegation took place
	MinTime          int64       `json:"min_time"`           // unix time for redelegation completion
	InitialBalance   sdk.Coin    `json:"initial_balance"`    // initial balance when redelegation started
	Balance          sdk.Coin    `json:"balance"`            // current balance
	SharesSrc        sdk.Rat     `json:"shares_src"`         // amount of source shares redelegating
	SharesDst        sdk.Rat     `json:"shares_dst"`         // amount of destination shares redelegating
}

// Equal returns a boolean determining if two Redelegation types are identical.
func (d Redelegation) Equal(d2 Redelegation) bool {
	bz1 := MsgCdc.MustMarshalBinary(&d)
	bz2 := MsgCdc.MustMarshalBinary(&d2)
	return bytes.Equal(bz1, bz2)
}

// HumanReadableString returns a human readable string representation of a
// Redelegation. An error is returned if the UnbondingDelegation's delegator or
// validator addresses cannot be Bech32 encoded.
func (d Redelegation) HumanReadableString() (string, error) {
	bechAcc, err := sdk.Bech32ifyAcc(d.DelegatorAddr)
	if err != nil {
		return "", err
	}

	bechValSrc, err := sdk.Bech32ifyAcc(d.ValidatorSrcAddr)
	if err != nil {
		return "", err
	}

	bechValDst, err := sdk.Bech32ifyAcc(d.ValidatorDstAddr)
	if err != nil {
		return "", err
	}

	resp := "Redelegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", bechAcc)
	resp += fmt.Sprintf("Source Validator: %s\n", bechValSrc)
	resp += fmt.Sprintf("Destination Validator: %s\n", bechValDst)
	resp += fmt.Sprintf("Creation height: %v\n", d.CreationHeight)
	resp += fmt.Sprintf("Min time to unbond (unix): %v\n", d.MinTime)
	resp += fmt.Sprintf("Source shares: %s", d.SharesSrc.String())
	resp += fmt.Sprintf("Destination shares: %s", d.SharesDst.String())

	return resp, nil

}
