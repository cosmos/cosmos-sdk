package keeper

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// gets an address from a validator's outstanding rewards key
func GetValidatorOutstandingRewardsAddress(key []byte) (valAddr sdk.ValAddress) {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.ValAddress(addr)
}

// gets an address from a delegator's withdraw info key
func GetDelegatorWithdrawInfoAddress(key []byte) (delAddr sdk.AccAddress) {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.AccAddress(addr)
}

// gets the addresses from a delegator starting info key
func GetDelegatorStartingInfoAddresses(key []byte) (valAddr sdk.ValAddress, delAddr sdk.AccAddress) {
	addr := key[1 : 1+sdk.AddrLen]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	valAddr = sdk.ValAddress(addr)
	addr = key[1+sdk.AddrLen:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	delAddr = sdk.AccAddress(addr)
	return
}

// gets the address & period from a validator's historical rewards key
func GetValidatorHistoricalRewardsAddressPeriod(key []byte) (valAddr sdk.ValAddress, period uint64) {
	addr := key[1 : 1+sdk.AddrLen]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	valAddr = sdk.ValAddress(addr)
	b := key[1+sdk.AddrLen:]
	if len(b) != 8 {
		panic("unexpected key length")
	}
	period = binary.LittleEndian.Uint64(b)
	return
}

// gets the address from a validator's current rewards key
func GetValidatorCurrentRewardsAddress(key []byte) (valAddr sdk.ValAddress) {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.ValAddress(addr)
}

// gets the address from a validator's accumulated commission key
func GetValidatorAccumulatedCommissionAddress(key []byte) (valAddr sdk.ValAddress) {
	addr := key[1:]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	return sdk.ValAddress(addr)
}

// gets the height from a validator's slash event key
func GetValidatorSlashEventAddressHeight(key []byte) (valAddr sdk.ValAddress, height uint64) {
	addr := key[1 : 1+sdk.AddrLen]
	if len(addr) != sdk.AddrLen {
		panic("unexpected key length")
	}
	valAddr = sdk.ValAddress(addr)
	startB := 1 + sdk.AddrLen
	b := key[startB : startB+8] // the next 8 bytes represent the height
	height = binary.BigEndian.Uint64(b)
	return
}

// gets the outstanding rewards key for a validator
func GetValidatorOutstandingRewardsKey(valAddr sdk.ValAddress) []byte {
	return append(types.ValidatorOutstandingRewardsPrefix, valAddr.Bytes()...)
}

// gets the key for a delegator's withdraw addr
func GetDelegatorWithdrawAddrKey(delAddr sdk.AccAddress) []byte {
	return append(types.DelegatorWithdrawAddrPrefix, delAddr.Bytes()...)
}

// gets the key for a delegator's starting info
func GetDelegatorStartingInfoKey(v sdk.ValAddress, d sdk.AccAddress) []byte {
	return append(append(types.DelegatorStartingInfoPrefix, v.Bytes()...), d.Bytes()...)
}

// gets the prefix key for a validator's historical rewards
func GetValidatorHistoricalRewardsPrefix(v sdk.ValAddress) []byte {
	return append(types.ValidatorHistoricalRewardsPrefix, v.Bytes()...)
}

// gets the key for a validator's historical rewards
func GetValidatorHistoricalRewardsKey(v sdk.ValAddress, k uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, k)
	return append(append(types.ValidatorHistoricalRewardsPrefix, v.Bytes()...), b...)
}

// gets the key for a validator's current rewards
func GetValidatorCurrentRewardsKey(v sdk.ValAddress) []byte {
	return append(types.ValidatorCurrentRewardsPrefix, v.Bytes()...)
}

// gets the key for a validator's current commission
func GetValidatorAccumulatedCommissionKey(v sdk.ValAddress) []byte {
	return append(types.ValidatorAccumulatedCommissionPrefix, v.Bytes()...)
}

// gets the prefix key for a validator's slash fractions
func GetValidatorSlashEventPrefix(v sdk.ValAddress) []byte {
	return append(types.ValidatorSlashEventPrefix, v.Bytes()...)
}

// gets the prefix key for a validator's slash fraction (ValidatorSlashEventPrefix + height)
func GetValidatorSlashEventKeyPrefix(v sdk.ValAddress, height uint64) []byte {
	heightBz := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBz, height)
	return append(
		types.ValidatorSlashEventPrefix,
		append(
			v.Bytes(),
			heightBz...,
		)...,
	)
}

// gets the key for a validator's slash fraction
func GetValidatorSlashEventKey(v sdk.ValAddress, height, period uint64) []byte {
	periodBz := make([]byte, 8)
	binary.BigEndian.PutUint64(periodBz, period)
	prefix := GetValidatorSlashEventKeyPrefix(v, height)
	return append(prefix, periodBz...)
}
