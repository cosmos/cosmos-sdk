package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/gogo/protobuf/proto"
)

// Keys for store prefixes
// Items are stored with the following keys:
//
// Params entry:
// - 0x00<name> -> <value>
// Sanctioned addresses:
// - 0x01<addr len (1 byte)><addr> -> 0x01
// Temporarily sanctioned or unsanctioned addresses:
// - 0x02<addr len (1 byte)><addr><gov prop id (8 bytes)> -> 0x01 or 0x00
// Proposal id temp sanction index:
// - 0x03<proposal id (8 bytes)><addr len (1 byte)><addr> -> 0x00 or 0x01
var (
	ParamsPrefix        = []byte{0x00}
	SanctionedPrefix    = []byte{0x01}
	TemporaryPrefix     = []byte{0x02}
	ProposalIndexPrefix = []byte{0x03}
)

const (
	ParamNameImmediateSanctionMinDeposit   = "immediate_sanction_min_deposit"
	ParamNameImmediateUnsanctionMinDeposit = "immediate_unsanction_min_deposit"
)

// ConcatBz creates a single byte slice consisting of the two provided byte slices.
// Like append() but always returns a new slice with its own underlying array.
func ConcatBz(bz1, bz2 []byte) []byte {
	return concatBzPlusCap(bz1, bz2, 0)
}

// concatBzPlusCap creates a single byte slice consisting of the two provided byte slices with some extra capacity in the underlying array.
// The idea is that you can append(...) to the result of this without it needed a new underlying array.
func concatBzPlusCap(bz1, bz2 []byte, extraCap int) []byte {
	l1 := len(bz1)
	l2 := len(bz2)
	rv := make([]byte, l1+l2, l1+l2+extraCap)
	if l1 > 0 {
		copy(rv, bz1)
	}
	if l2 > 0 {
		copy(rv[l1:], bz2)
	}
	return rv
}

// ParseLengthPrefixedBz parses a length-prefixed byte slice into those bytes and any leftover bytes.
func ParseLengthPrefixedBz(bz []byte) ([]byte, []byte) {
	addrLen, addrLenEndIndex := sdk.ParseLengthPrefixedBytes(bz, 0, 1)
	addr, addrEndIndex := sdk.ParseLengthPrefixedBytes(bz, addrLenEndIndex+1, int(addrLen[0]))
	var remainder []byte
	if len(bz) > addrEndIndex+1 {
		remainder = bz[addrEndIndex+1:]
	}
	return addr, remainder
}

// CreateParamKey creates the key to use for a param with the given name.
//
// - 0x00<name> -> <value>
func CreateParamKey(name string) []byte {
	return ConcatBz(ParamsPrefix, []byte(name))
}

// ParseParamKey extracts the param name from the provided key.
func ParseParamKey(bz []byte) string {
	return string(bz[1:])
}

// CreateSanctionedAddrKey creates the sanctioned address key for the provided address.
//
// - 0x01<addr len (1 byte)><addr>
func CreateSanctionedAddrKey(addr sdk.AccAddress) []byte {
	return ConcatBz(SanctionedPrefix, address.MustLengthPrefix(addr))
}

// ParseSanctionedAddrKey extracts the address from the provided sanctioned address key.
func ParseSanctionedAddrKey(key []byte) sdk.AccAddress {
	addr, _ := ParseLengthPrefixedBz(key[1:])
	return addr
}

// CreateTemporaryAddrPrefix creates a key prefix for a temporarily sanctioned/unsanctioned address.
//
// If an address is provided:
// - 0x02<addr len(1 byte)><addr>
// If an address isn't provided:
// - 0x02
func CreateTemporaryAddrPrefix(addr sdk.AccAddress) []byte {
	if len(addr) == 0 {
		return ConcatBz(TemporaryPrefix, []byte{})
	}
	return concatBzPlusCap(TemporaryPrefix, address.MustLengthPrefix(addr), 8)
}

// CreateTemporaryKey creates a key for a temporarily sanctioned/unsanctioned address associated with the given governance proposal id.
//
// - 0x02<addr len (1 byte)><addr><gov prop id (8 bytes)>
func CreateTemporaryKey(addr sdk.AccAddress, govPropID uint64) []byte {
	pre := CreateTemporaryAddrPrefix(addr)
	idBz := sdk.Uint64ToBigEndian(govPropID)
	if len(pre)+8 == cap(pre) {
		return append(CreateTemporaryAddrPrefix(addr), idBz...)
	}
	return ConcatBz(pre, idBz)
}

// ParseTemporaryKey extracts the address and gov prop id from the provided temporary key.
func ParseTemporaryKey(key []byte) (sdk.AccAddress, uint64) {
	addr, govPropIDBz := ParseLengthPrefixedBz(key[1:])
	govPropID := sdk.BigEndianToUint64(govPropIDBz)
	return addr, govPropID
}

const (
	// SanctionB is a byte representing a sanction (either temporary or permanent).
	SanctionB = 0x01
	// UnsanctionB is a byte representing an unsanction (probably temporary).
	UnsanctionB = 0x00
)

// IsSanctionBz returns true if the provided byte slice indicates a temporary sanction.
func IsSanctionBz(bz []byte) bool {
	return len(bz) == 1 && bz[0] == SanctionB
}

// IsUnsanctionBz returns true if the provided byte slice indicates a temporary unsanction.
func IsUnsanctionBz(bz []byte) bool {
	return len(bz) == 1 && bz[0] == UnsanctionB
}

// ToTempStatus converts a temporary entry value byte slice into a TempStatus value.
func ToTempStatus(bz []byte) sanction.TempStatus {
	if len(bz) == 1 {
		switch bz[0] {
		case SanctionB:
			return sanction.TEMP_STATUS_SANCTIONED
		case UnsanctionB:
			return sanction.TEMP_STATUS_UNSANCTIONED
		}
	}
	return sanction.TEMP_STATUS_UNSPECIFIED
}

// NewTempEvent creates the temp event for the given type val (e.g. SanctionB or UnsanctionB) with the given address.
func NewTempEvent(typeVal byte, addr sdk.AccAddress) proto.Message {
	switch typeVal {
	case SanctionB:
		return sanction.NewEventTempAddressSanctioned(addr)
	case UnsanctionB:
		return sanction.NewEventTempAddressUnsanctioned(addr)
	default:
		panic(fmt.Errorf("unknown temp value byte: %x", typeVal))
	}
}

// CreateProposalTempIndexPrefix creates a key prefix for a proposal temporary index key.
//
// If a govPropID is provided:
// - 0x03<proposal id (8 bytes)>
// If a govPropID isn't provided:
// - 0x03
func CreateProposalTempIndexPrefix(govPropID *uint64) []byte {
	if govPropID == nil {
		return ConcatBz(ProposalIndexPrefix, []byte{})
	}
	return concatBzPlusCap(ProposalIndexPrefix, sdk.Uint64ToBigEndian(*govPropID), 33)
}

// CreateProposalTempIndexKey creates a key for a proposal id + addr temporary index entry.
//
// 0x03<proposal id (8 bytes)><addr len (1 byte)><addr>
func CreateProposalTempIndexKey(govPropID uint64, addr sdk.AccAddress) []byte {
	return append(CreateProposalTempIndexPrefix(&govPropID), address.MustLengthPrefix(addr)...)
}

// ParseProposalTempIndexKey extracts the gov prop id and address from the provided proposal temp index key.
func ParseProposalTempIndexKey(key []byte) (uint64, sdk.AccAddress) {
	govPropID := sdk.BigEndianToUint64(key[1:9])
	addr, _ := ParseLengthPrefixedBz(key[9:])
	return govPropID, addr
}
