package quarantine

import (
	"bytes"
	"crypto/sha256"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// ModuleName is the name of the module
	ModuleName = "quarantine"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName
)

var (
	// OptInPrefix is the prefix for the quarantine account opt-in flags.
	OptInPrefix = []byte{0x00}

	// AutoResponsePrefix is the prefix for quarantine auto-response settings.
	AutoResponsePrefix = []byte{0x01}

	// RecordPrefix is the prefix for keys with the records of quarantined funds.
	RecordPrefix = []byte{0x02}

	// RecordIndexPrefix is the prefix for the index of record suffixes.
	RecordIndexPrefix = []byte{0x03}
)

// MakeKey concatenates the two byte slices into a new byte slice.
func MakeKey(part1, part2 []byte) []byte {
	rv := make([]byte, len(part1)+len(part2))
	copy(rv, part1)
	copy(rv[len(part1):], part2)
	return rv
}

// CreateOptInKey creates the key for a quarantine opt-in record.
func CreateOptInKey(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	return MakeKey(OptInPrefix, toAddrBz)
}

// ParseOptInKey extracts the account address from the provided quarantine opt-in key.
func ParseOptInKey(key []byte) (toAddr sdk.AccAddress) {
	// key is of format:
	// 0x00<to addr len><to addr bytes>
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, _ = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	return toAddr
}

// CreateAutoResponseToAddrPrefix creates a prefix for the quarantine auto-responses for a receiving address.
func CreateAutoResponseToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	return MakeKey(AutoResponsePrefix, toAddrBz)
}

// CreateAutoResponseKey creates the key for a quarantine auto-response.
func CreateAutoResponseKey(toAddr, fromAddr sdk.AccAddress) []byte {
	toAddrPreBz := CreateAutoResponseToAddrPrefix(toAddr)
	fromAddrBz := address.MustLengthPrefix(fromAddr)
	return MakeKey(toAddrPreBz, fromAddrBz)
}

// ParseAutoResponseKey extracts the to address and from address from the provided quarantine auto-response key.
func ParseAutoResponseKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x01<to addr len><to addr bytes><from addr len><from addr bytes>
	var toAddrEndIndex int
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}

// CreateRecordToAddrPrefix creates a prefix for the quarantine funds for a receiving address.
func CreateRecordToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	return MakeKey(RecordPrefix, toAddrBz)
}

// CreateRecordKey creates the key for a quarantine record.
//
// If there is only one fromAddr, it is used as the record suffix.
// If there are more than one, a hash of them is used as the suffix.
//
// Panics if no fromAddrs are provided.
func CreateRecordKey(toAddr sdk.AccAddress, fromAddrs ...sdk.AccAddress) []byte {
	// This is designed such that a known record suffix can be provided
	// as a single "from address" to create the key with that suffix.
	toAddrPreBz := CreateRecordToAddrPrefix(toAddr)
	recordId := address.MustLengthPrefix(createRecordSuffix(fromAddrs))
	return MakeKey(toAddrPreBz, recordId)
}

// createRecordSuffix creates a single "address" to use for the provided from addresses.
// This is not to be confused with CreateRecordKey which creates the full key for a quarantine record.
// This only creates a portion of the key.
//
// If one fromAddr is provided, it's what's returned.
// If more than one is provided, they are sorted, combined, and hashed.
//
// Panics if none are provided.
func createRecordSuffix(fromAddrs []sdk.AccAddress) []byte {
	// This is designed such that a known record suffix can be provided
	// as a single "from address" to create the key with that suffix.
	switch len(fromAddrs) {
	case 0:
		panic(sdkerrors.ErrLogic.Wrap("at least one fromAddr is required"))
	case 1:
		if len(fromAddrs[0]) > 32 {
			return fromAddrs[0][:32]
		}
		return fromAddrs[0]
	default:
		// The same n addresses needs to always create the same result.
		// And we don't want to change the input slice.
		addrs := make([]sdk.AccAddress, len(fromAddrs))
		copy(addrs, fromAddrs)
		sort.Slice(addrs, func(i, j int) bool {
			return bytes.Compare(addrs[i], addrs[j]) < 0
		})
		var toHash []byte
		for _, addr := range addrs {
			toHash = append(toHash, addr...)
		}
		hash := sha256.Sum256(toHash)
		return hash[0:]
	}
}

// ParseRecordKey extracts the to address and record suffix from the provided quarantine funds key.
func ParseRecordKey(key []byte) (toAddr, recordSuffix sdk.AccAddress) {
	// key is of format:
	// 0x02<to addr len><to addr bytes><record suffix len><record suffix bytes>
	var toAddrEndIndex int
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	recordSuffixLen, recordSuffixLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	recordSuffix, _ = sdk.ParseLengthPrefixedBytes(key, recordSuffixLenEndIndex+1, int(recordSuffixLen[0]))

	return toAddr, recordSuffix
}

// CreateRecordIndexToAddrPrefix creates a prefix for the quarantine record index entries for a receiving address.
func CreateRecordIndexToAddrPrefix(toAddr sdk.AccAddress) []byte {
	toAddrBz := address.MustLengthPrefix(toAddr)
	return MakeKey(RecordIndexPrefix, toAddrBz)
}

// CreateRecordIndexKey creates the key for the quarantine record suffix index.
func CreateRecordIndexKey(toAddr, fromAddr sdk.AccAddress) []byte {
	toAddrPreBz := CreateRecordIndexToAddrPrefix(toAddr)
	recordId := address.MustLengthPrefix(fromAddr)
	return MakeKey(toAddrPreBz, recordId)
}

// ParseRecordIndexKey extracts the to address and from address from the provided quarantine record index key.
func ParseRecordIndexKey(key []byte) (toAddr, fromAddr sdk.AccAddress) {
	// key is of format:
	// 0x03<to addr len><to addr bytes><from addr len><from addr bytes>
	var toAddrEndIndex int
	toAddrLen, toAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, 1, 1)
	toAddr, toAddrEndIndex = sdk.ParseLengthPrefixedBytes(key, toAddrLenEndIndex+1, int(toAddrLen[0]))

	fromAddrLen, fromAddrLenEndIndex := sdk.ParseLengthPrefixedBytes(key, toAddrEndIndex+1, 1)
	fromAddr, _ = sdk.ParseLengthPrefixedBytes(key, fromAddrLenEndIndex+1, int(fromAddrLen[0]))

	return toAddr, fromAddr
}
