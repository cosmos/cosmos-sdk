package types

import (
	"bytes"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Delegation represents the bond with tokens held by an account.  It is
// owned by one delegator, and is associated with the voting power of one
// pubKey.
type Delegation struct {
	DelegatorAddr sdk.AccAddress `json:"delegator_addr"`
	ValidatorAddr sdk.AccAddress `json:"validator_addr"`
	Shares        sdk.Rat        `json:"shares"`
	Height        int64          `json:"height"` // Last height bond updated
}

type delegationValue struct {
	Shares sdk.Rat
	Height int64
}

// return the delegation without fields contained within the key for the store
func MustMarshalDelegation(cdc *wire.Codec, delegation Delegation) []byte {
	val := delegationValue{
		delegation.Shares,
		delegation.Height,
	}
	return cdc.MustMarshalBinary(val)
}

// return the delegation without fields contained within the key for the store
func MustUnmarshalDelegation(cdc *wire.Codec, key, value []byte) Delegation {
	delegation, err := UnmarshalDelegation(cdc, key, value)
	if err != nil {
		panic(err)
	}
	return delegation
}

// return the delegation without fields contained within the key for the store
func UnmarshalDelegation(cdc *wire.Codec, key, value []byte) (delegation Delegation, err error) {
	var storeValue delegationValue
	err = cdc.UnmarshalBinary(value, &storeValue)
	if err != nil {
		return
	}

	addrs := key[1:] // remove prefix bytes
	if len(addrs) != 2*sdk.AddrLen {
		err = errors.New("unexpected key length")
		return
	}
	delAddr := sdk.AccAddress(addrs[:sdk.AddrLen])
	valAddr := sdk.AccAddress(addrs[sdk.AddrLen:])

	return Delegation{
		DelegatorAddr: delAddr,
		ValidatorAddr: valAddr,
		Shares:        storeValue.Shares,
		Height:        storeValue.Height,
	}, nil
}

// nolint
func (d Delegation) Equal(d2 Delegation) bool {
	return bytes.Equal(d.DelegatorAddr, d2.DelegatorAddr) &&
		bytes.Equal(d.ValidatorAddr, d2.ValidatorAddr) &&
		d.Height == d2.Height &&
		d.Shares.Equal(d2.Shares)
}

// ensure fulfills the sdk validator types
var _ sdk.Delegation = Delegation{}

// nolint - for sdk.Delegation
func (d Delegation) GetDelegator() sdk.AccAddress { return d.DelegatorAddr }
func (d Delegation) GetValidator() sdk.AccAddress { return d.ValidatorAddr }
func (d Delegation) GetBondShares() sdk.Rat       { return d.Shares }

// HumanReadableString returns a human readable string representation of a
// Delegation. An error is returned if the Delegation's delegator or validator
// addresses cannot be Bech32 encoded.
func (d Delegation) HumanReadableString() (string, error) {
	resp := "Delegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", d.DelegatorAddr)
	resp += fmt.Sprintf("Validator: %s\n", d.ValidatorAddr)
	resp += fmt.Sprintf("Shares: %s", d.Shares.String())
	resp += fmt.Sprintf("Height: %d", d.Height)

	return resp, nil
}

// UnbondingDelegation reflects a delegation's passive unbonding queue.
type UnbondingDelegation struct {
	DelegatorAddr  sdk.AccAddress `json:"delegator_addr"`  // delegator
	ValidatorAddr  sdk.AccAddress `json:"validator_addr"`  // validator unbonding from owner addr
	CreationHeight int64          `json:"creation_height"` // height which the unbonding took place
	MinTime        int64          `json:"min_time"`        // unix time for unbonding completion
	InitialBalance sdk.Coin       `json:"initial_balance"` // atoms initially scheduled to receive at completion
	Balance        sdk.Coin       `json:"balance"`         // atoms to receive at completion
}

type ubdValue struct {
	CreationHeight int64
	MinTime        int64
	InitialBalance sdk.Coin
	Balance        sdk.Coin
}

// return the unbonding delegation without fields contained within the key for the store
func MustMarshalUBD(cdc *wire.Codec, ubd UnbondingDelegation) []byte {
	val := ubdValue{
		ubd.CreationHeight,
		ubd.MinTime,
		ubd.InitialBalance,
		ubd.Balance,
	}
	return cdc.MustMarshalBinary(val)
}

// unmarshal a unbonding delegation from a store key and value
func MustUnmarshalUBD(cdc *wire.Codec, key, value []byte) UnbondingDelegation {
	ubd, err := UnmarshalUBD(cdc, key, value)
	if err != nil {
		panic(err)
	}
	return ubd
}

// unmarshal a unbonding delegation from a store key and value
func UnmarshalUBD(cdc *wire.Codec, key, value []byte) (ubd UnbondingDelegation, err error) {
	var storeValue ubdValue
	err = cdc.UnmarshalBinary(value, &storeValue)
	if err != nil {
		return
	}

	addrs := key[1:] // remove prefix bytes
	if len(addrs) != 2*sdk.AddrLen {
		err = errors.New("unexpected key length")
		return
	}
	delAddr := sdk.AccAddress(addrs[:sdk.AddrLen])
	valAddr := sdk.AccAddress(addrs[sdk.AddrLen:])

	return UnbondingDelegation{
		DelegatorAddr:  delAddr,
		ValidatorAddr:  valAddr,
		CreationHeight: storeValue.CreationHeight,
		MinTime:        storeValue.MinTime,
		InitialBalance: storeValue.InitialBalance,
		Balance:        storeValue.Balance,
	}, nil
}

// nolint
func (d UnbondingDelegation) Equal(d2 UnbondingDelegation) bool {
	bz1 := MsgCdc.MustMarshalBinary(&d)
	bz2 := MsgCdc.MustMarshalBinary(&d2)
	return bytes.Equal(bz1, bz2)
}

// HumanReadableString returns a human readable string representation of an
// UnbondingDelegation. An error is returned if the UnbondingDelegation's
// delegator or validator addresses cannot be Bech32 encoded.
func (d UnbondingDelegation) HumanReadableString() (string, error) {
	resp := "Unbonding Delegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", d.DelegatorAddr)
	resp += fmt.Sprintf("Validator: %s\n", d.ValidatorAddr)
	resp += fmt.Sprintf("Creation height: %v\n", d.CreationHeight)
	resp += fmt.Sprintf("Min time to unbond (unix): %v\n", d.MinTime)
	resp += fmt.Sprintf("Expected balance: %s", d.Balance.String())

	return resp, nil

}

// Redelegation reflects a delegation's passive re-delegation queue.
type Redelegation struct {
	DelegatorAddr    sdk.AccAddress `json:"delegator_addr"`     // delegator
	ValidatorSrcAddr sdk.AccAddress `json:"validator_src_addr"` // validator redelegation source owner addr
	ValidatorDstAddr sdk.AccAddress `json:"validator_dst_addr"` // validator redelegation destination owner addr
	CreationHeight   int64          `json:"creation_height"`    // height which the redelegation took place
	MinTime          int64          `json:"min_time"`           // unix time for redelegation completion
	InitialBalance   sdk.Coin       `json:"initial_balance"`    // initial balance when redelegation started
	Balance          sdk.Coin       `json:"balance"`            // current balance
	SharesSrc        sdk.Rat        `json:"shares_src"`         // amount of source shares redelegating
	SharesDst        sdk.Rat        `json:"shares_dst"`         // amount of destination shares redelegating
}

type redValue struct {
	CreationHeight int64
	MinTime        int64
	InitialBalance sdk.Coin
	Balance        sdk.Coin
	SharesSrc      sdk.Rat
	SharesDst      sdk.Rat
}

// return the redelegation without fields contained within the key for the store
func MustMarshalRED(cdc *wire.Codec, red Redelegation) []byte {
	val := redValue{
		red.CreationHeight,
		red.MinTime,
		red.InitialBalance,
		red.Balance,
		red.SharesSrc,
		red.SharesDst,
	}
	return cdc.MustMarshalBinary(val)
}

// unmarshal a redelegation from a store key and value
func MustUnmarshalRED(cdc *wire.Codec, key, value []byte) Redelegation {
	red, err := UnmarshalRED(cdc, key, value)
	if err != nil {
		panic(err)
	}
	return red
}

// unmarshal a redelegation from a store key and value
func UnmarshalRED(cdc *wire.Codec, key, value []byte) (red Redelegation, err error) {
	var storeValue redValue
	err = cdc.UnmarshalBinary(value, &storeValue)
	if err != nil {
		return
	}

	addrs := key[1:] // remove prefix bytes
	if len(addrs) != 3*sdk.AddrLen {
		err = errors.New("unexpected key length")
		return
	}
	delAddr := sdk.AccAddress(addrs[:sdk.AddrLen])
	valSrcAddr := sdk.AccAddress(addrs[sdk.AddrLen : 2*sdk.AddrLen])
	valDstAddr := sdk.AccAddress(addrs[2*sdk.AddrLen:])

	return Redelegation{
		DelegatorAddr:    delAddr,
		ValidatorSrcAddr: valSrcAddr,
		ValidatorDstAddr: valDstAddr,
		CreationHeight:   storeValue.CreationHeight,
		MinTime:          storeValue.MinTime,
		InitialBalance:   storeValue.InitialBalance,
		Balance:          storeValue.Balance,
		SharesSrc:        storeValue.SharesSrc,
		SharesDst:        storeValue.SharesDst,
	}, nil
}

// nolint
func (d Redelegation) Equal(d2 Redelegation) bool {
	bz1 := MsgCdc.MustMarshalBinary(&d)
	bz2 := MsgCdc.MustMarshalBinary(&d2)
	return bytes.Equal(bz1, bz2)
}

// HumanReadableString returns a human readable string representation of a
// Redelegation. An error is returned if the UnbondingDelegation's delegator or
// validator addresses cannot be Bech32 encoded.
func (d Redelegation) HumanReadableString() (string, error) {
	resp := "Redelegation \n"
	resp += fmt.Sprintf("Delegator: %s\n", d.DelegatorAddr)
	resp += fmt.Sprintf("Source Validator: %s\n", d.ValidatorSrcAddr)
	resp += fmt.Sprintf("Destination Validator: %s\n", d.ValidatorDstAddr)
	resp += fmt.Sprintf("Creation height: %v\n", d.CreationHeight)
	resp += fmt.Sprintf("Min time to unbond (unix): %v\n", d.MinTime)
	resp += fmt.Sprintf("Source shares: %s", d.SharesSrc.String())
	resp += fmt.Sprintf("Destination shares: %s", d.SharesDst.String())

	return resp, nil

}
