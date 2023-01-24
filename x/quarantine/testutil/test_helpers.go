package testutil

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
)

// This file contains some functions handy for doing unit tests.

// AssertErrorContents asserts that, if contains is empty, there's no error.
// Otherwise, asserts that there is an error, and that it contains each of the provided strings.
func AssertErrorContents(t *testing.T, theError error, contains []string, msgAndArgs ...interface{}) bool {
	t.Helper()
	if len(contains) == 0 {
		return assert.NoError(t, theError, msgAndArgs)
	}
	rv := assert.Error(t, theError, msgAndArgs...)
	if rv {
		for _, expInErr := range contains {
			rv = assert.ErrorContains(t, theError, expInErr, msgAndArgs...) && rv
		}
	}
	return rv
}

// MakeTestAddr makes an AccAddress that's 20 bytes long.
// The first byte is the index. The next bytes are the base.
// The byte after that is 97 (a) + the index.
// Each other byte is one more than the previous.
// Panics if the base is too long or index too high.
func MakeTestAddr(base string, index uint8) sdk.AccAddress {
	return makePrefixedIncAddr(20, 97, base, index)
}

// MakeLongAddr makes an AccAddress that's 32 bytes long.
// The first byte is the index. The next bytes are the base.
// The byte after that is 65 (A) + the index.
// Each other byte is one more than the previous.
// Panics if the base is too long or index too high.
func MakeLongAddr(base string, index uint8) sdk.AccAddress {
	return makePrefixedIncAddr(32, 65, base, index)
}

// MakeBadAddr makes an address that's longer than the max length allowed.
// The first byte is the index. The next bytes are the base.
// The byte after that is 33 (!) + the index.
// Each other byte is one more than the previous wrapping back to 0 after 255.
// Panics if the base is too long or index too high.
func MakeBadAddr(base string, index uint8) sdk.AccAddress {
	return makePrefixedIncAddr(address.MaxAddrLen+1, 33, base, index)
}

// makePrefixedIncAddr creates an sdk.AccAddress with the provided length.
// The first byte will be the index.
// The next bytes will be the base.
// The byte after that will be the rootChar+index.
// Each other byte is one more than the previous, skipping 127, and wrapping back to 33 after 254.
// Panics if the base is longer than 8 chars. Keep it short. There aren't that many bytes to work with here.
// Panics if the index is larger than 30. You don't need that many.
// Panics if rootChar+index is not 33 to 126 or 128 to 254 (inclusive). Keep them printable.
//
// You're probably looking for MakeTestAddr, MakeLongAddr, or MakeBadAddr.
func makePrefixedIncAddr(length uint, rootChar uint8, base string, index uint8) sdk.AccAddress {
	// panics are used because this is for test setup and should be mostly hard-coded stuff anyway.
	// 8 is picked because if the length is 20, that would only leave 11 more bytes, which isn't many.
	if len(base) > 8 {
		panic(fmt.Sprintf("base too long %q; got: %d, max: 8", base, len(base)))
	}
	// 25 is picked so the long and test addresses always start with a letter.
	if index > 25 {
		panic(fmt.Sprintf("index too large; got: %d, max: 30", index))
	}
	rv := makeIncAddr(length, uint(1+len(base))+1, rootChar+index)
	rv[0] = index
	copy(rv[1:], base)
	return rv
}

// makeIncAddr creates an sdk.AccAddress with the provided length.
// The first firstCharLen bytes will be the firstChar.
// Each other byte is one more than the previous, skipping 127, and wrapping back to 33 after 254.
// Basically using only bytes that are printable as ascii.
// If the firstChar is anything other than 33 to 126 or 128 to 254 (inclusive), you're gonna have a bad time.
//
// You're probably looking for MakeTestAddr, MakeLongAddr, or MakeBadAddr.
func makeIncAddr(length, firstCharLen uint, firstChar uint8) sdk.AccAddress {
	// At one point I used math and mod stuff to wrap the provided first char into the right range,
	// but this is for tests and should be from mostly hard-coded stuff anyway, so panics are used.
	if firstChar < 33 || firstChar > 254 || firstChar == 127 {
		panic(fmt.Sprintf("illegal shift (5-yard penalty, retry down): expected 33-126 or 128-254, got: %d", firstChar))
	}
	b := firstChar
	rv := make(sdk.AccAddress, length)
	for i := uint(0); i < length; i++ {
		if i >= firstCharLen {
			switch {
			case b == 126:
				b = 128
			case b >= 254:
				b = 33
			default:
				b++
			}
		}
		rv[i] = b
	}
	return rv
}

// MakeCopyOfCoins makes a deep copy of some Coins.
func MakeCopyOfCoins(orig sdk.Coins) sdk.Coins {
	if orig == nil {
		return nil
	}
	rv := make(sdk.Coins, len(orig))
	for i, coin := range orig {
		rv[i] = sdk.Coin{
			Denom:  coin.Denom,
			Amount: coin.Amount.AddRaw(0),
		}
	}
	return rv
}

// MakeCopyOfQuarantinedFunds makes a deep copy of a QuarantinedFunds.
func MakeCopyOfQuarantinedFunds(orig *quarantine.QuarantinedFunds) *quarantine.QuarantinedFunds {
	return &quarantine.QuarantinedFunds{
		ToAddress:               orig.ToAddress,
		UnacceptedFromAddresses: MakeCopyOfStringSlice(orig.UnacceptedFromAddresses),
		Coins:                   MakeCopyOfCoins(orig.Coins),
		Declined:                orig.Declined,
	}
}

// MakeCopyOfQuarantinedFundsSlice makes a deep copy of a slice of QuarantinedFunds.
func MakeCopyOfQuarantinedFundsSlice(orig []*quarantine.QuarantinedFunds) []*quarantine.QuarantinedFunds {
	if orig == nil {
		return orig
	}
	rv := make([]*quarantine.QuarantinedFunds, len(orig))
	for i, qf := range orig {
		rv[i] = MakeCopyOfQuarantinedFunds(qf)
	}
	return rv
}

// MakeCopyOfStringSlice makes a deep copy of a slice of strings.
func MakeCopyOfStringSlice(orig []string) []string {
	if orig == nil {
		return nil
	}
	rv := make([]string, len(orig))
	copy(rv, orig)
	return rv
}

// MakeCopyOfQuarantineRecord makes a deep copy of a QuarantineRecord.
func MakeCopyOfQuarantineRecord(orig *quarantine.QuarantineRecord) *quarantine.QuarantineRecord {
	return &quarantine.QuarantineRecord{
		UnacceptedFromAddresses: MakeCopyOfAccAddresses(orig.UnacceptedFromAddresses),
		AcceptedFromAddresses:   MakeCopyOfAccAddresses(orig.AcceptedFromAddresses),
		Coins:                   MakeCopyOfCoins(orig.Coins),
		Declined:                orig.Declined,
	}
}

// MakeCopyOfAccAddress makes a deep copy of an AccAddress.
func MakeCopyOfAccAddress(orig sdk.AccAddress) sdk.AccAddress {
	if orig == nil {
		return orig
	}
	rv := make(sdk.AccAddress, len(orig))
	copy(rv, orig)
	return rv
}

// MakeCopyOfAccAddresses makes a deep copy of a slice of AccAddresses.
func MakeCopyOfAccAddresses(orig []sdk.AccAddress) []sdk.AccAddress {
	if orig == nil {
		return nil
	}
	rv := make([]sdk.AccAddress, len(orig))
	for i, addr := range orig {
		rv[i] = MakeCopyOfAccAddress(addr)
	}
	return rv
}

// MakeCopyOfByteSlice makes a deep copy of a byte slice.
func MakeCopyOfByteSlice(orig []byte) []byte {
	if orig == nil {
		return nil
	}
	rv := make([]byte, len(orig))
	copy(rv, orig)
	return rv
}

// MakeCopyOfByteSliceSlice makes a deep copy of a slice of byte slices.
func MakeCopyOfByteSliceSlice(orig [][]byte) [][]byte {
	if orig == nil {
		return nil
	}
	rv := make([][]byte, len(orig))
	for i, bz := range orig {
		rv[i] = MakeCopyOfByteSlice(bz)
	}
	return rv
}

// MakeCopyOfGenesisState makes a deep copy of a GenesisState.
func MakeCopyOfGenesisState(orig *quarantine.GenesisState) *quarantine.GenesisState {
	if orig == nil {
		return nil
	}
	return &quarantine.GenesisState{
		QuarantinedAddresses: MakeCopyOfStringSlice(orig.QuarantinedAddresses),
		AutoResponses:        MakeCopyOfAutoResponseEntries(orig.AutoResponses),
		QuarantinedFunds:     MakeCopyOfQuarantinedFundsSlice(orig.QuarantinedFunds),
	}
}

// MakeCopyOfAutoResponseEntries makes a deep copy of a slice of AutoResponseEntries.
func MakeCopyOfAutoResponseEntries(orig []*quarantine.AutoResponseEntry) []*quarantine.AutoResponseEntry {
	if orig == nil {
		return nil
	}
	rv := make([]*quarantine.AutoResponseEntry, len(orig))
	for i, entry := range orig {
		rv[i] = MakeCopyOfAutoResponseEntry(entry)
	}
	return rv
}

// MakeCopyOfAutoResponseEntry makes a deep copy of an AutoResponseEntry.
func MakeCopyOfAutoResponseEntry(orig *quarantine.AutoResponseEntry) *quarantine.AutoResponseEntry {
	if orig == nil {
		return nil
	}
	return &quarantine.AutoResponseEntry{
		ToAddress:   orig.ToAddress,
		FromAddress: orig.FromAddress,
		Response:    orig.Response,
	}
}

// MakeCopyOfQuarantineRecordSuffixIndex makes a deep copy of a QuarantineRecordSuffixIndex
func MakeCopyOfQuarantineRecordSuffixIndex(orig *quarantine.QuarantineRecordSuffixIndex) *quarantine.QuarantineRecordSuffixIndex {
	if orig == nil {
		return nil
	}
	return &quarantine.QuarantineRecordSuffixIndex{
		RecordSuffixes: MakeCopyOfByteSliceSlice(orig.RecordSuffixes),
	}
}
