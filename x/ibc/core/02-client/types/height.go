package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var _ exported.Height = (*Height)(nil)

// IsVersionFormat checks if a chainID is in the format required for parsing versions
// The chainID must be in the form: `{chainID}-{version}
// 24-host may enforce stricter checks on chainID
var IsVersionFormat = regexp.MustCompile(`^.+[^-]-{1}[1-9][0-9]*$`).MatchString

// ZeroHeight is a helper function which returns an uninitialized height.
func ZeroHeight() Height {
	return Height{}
}

// NewHeight is a constructor for the IBC height type
func NewHeight(versionNumber, versionHeight uint64) Height {
	return Height{
		VersionNumber: versionNumber,
		VersionHeight: versionHeight,
	}
}

// GetVersionNumber returns the version-number of the height
func (h Height) GetVersionNumber() uint64 {
	return h.VersionNumber
}

// GetVersionHeight returns the version-height of the height
func (h Height) GetVersionHeight() uint64 {
	return h.VersionHeight
}

// Compare implements a method to compare two heights. When comparing two heights a, b
// we can call a.Compare(b) which will return
// -1 if a < b
// 0  if a = b
// 1  if a > b
//
// It first compares based on version numbers, whichever has the higher version number is the higher height
// If version number is the same, then the version height is compared
func (h Height) Compare(other exported.Height) int64 {
	height, ok := other.(Height)
	if !ok {
		panic(fmt.Sprintf("cannot compare against invalid height type: %T. expected height type: %T", other, h))
	}
	var cmp int64
	if h.VersionNumber != height.VersionNumber {
		cmp = int64(h.VersionNumber) - int64(height.VersionNumber)
	} else {
		cmp = int64(h.VersionHeight) - int64(height.VersionHeight)
	}
	if cmp < 0 {
		return -1
	} else if cmp > 0 {
		return 1
	}
	return 0
}

// LT Helper comparison function returns true if h < other
func (h Height) LT(other exported.Height) bool {
	return h.Compare(other) == -1
}

// LTE Helper comparison function returns true if h <= other
func (h Height) LTE(other exported.Height) bool {
	cmp := h.Compare(other)
	return cmp <= 0
}

// GT Helper comparison function returns true if h > other
func (h Height) GT(other exported.Height) bool {
	return h.Compare(other) == 1
}

// GTE Helper comparison function returns true if h >= other
func (h Height) GTE(other exported.Height) bool {
	cmp := h.Compare(other)
	return cmp >= 0
}

// EQ Helper comparison function returns true if h == other
func (h Height) EQ(other exported.Height) bool {
	return h.Compare(other) == 0
}

// String returns a string representation of Height
func (h Height) String() string {
	return fmt.Sprintf("%d-%d", h.VersionNumber, h.VersionHeight)
}

// Decrement will return a new height with the VersionHeight decremented
// If the VersionHeight is already at lowest value (1), then false success flag is returend
func (h Height) Decrement() (decremented exported.Height, success bool) {
	if h.VersionHeight == 0 {
		return Height{}, false
	}
	return NewHeight(h.VersionNumber, h.VersionHeight-1), true
}

// Increment will return a height with the same version number but an
// incremented version height
func (h Height) Increment() Height {
	return NewHeight(h.VersionNumber, h.VersionHeight+1)
}

// IsZero returns true if height version and version-height are both 0
func (h Height) IsZero() bool {
	return h.VersionNumber == 0 && h.VersionHeight == 0
}

// MustParseHeight will attempt to parse a string representation of a height and panic if
// parsing fails.
func MustParseHeight(heightStr string) Height {
	height, err := ParseHeight(heightStr)
	if err != nil {
		panic(err)
	}

	return height
}

// ParseHeight is a utility function that takes a string representation of the height
// and returns a Height struct
func ParseHeight(heightStr string) (Height, error) {
	splitStr := strings.Split(heightStr, "-")
	if len(splitStr) != 2 {
		return Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "expected height string format: {version}-{height}. Got: %s", heightStr)
	}
	versionNumber, err := strconv.ParseUint(splitStr[0], 10, 64)
	if err != nil {
		return Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid version number. parse err: %s", err)
	}
	versionHeight, err := strconv.ParseUint(splitStr[1], 10, 64)
	if err != nil {
		return Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid version height. parse err: %s", err)
	}
	return NewHeight(versionNumber, versionHeight), nil
}

// SetVersionNumber takes a chainID in valid version format and swaps the version number
// in the chainID with the given version number.
func SetVersionNumber(chainID string, version uint64) (string, error) {
	if !IsVersionFormat(chainID) {
		return "", sdkerrors.Wrapf(
			sdkerrors.ErrInvalidChainID, "chainID is not in version format: %s", chainID,
		)
	}

	splitStr := strings.Split(chainID, "-")
	// swap out version number with given version
	splitStr[len(splitStr)-1] = strconv.Itoa(int(version))
	return strings.Join(splitStr, "-"), nil
}

// ParseChainID is a utility function that returns an version number from the given ChainID.
// ParseChainID attempts to parse a chain id in the format: `{chainID}-{version}`
// and return the versionnumber as a uint64.
// If the chainID is not in the expected format, a default version value of 0 is returned.
func ParseChainID(chainID string) uint64 {
	if !IsVersionFormat(chainID) {
		// chainID is not in version format, return 0 as default
		return 0
	}
	splitStr := strings.Split(chainID, "-")
	version, err := strconv.ParseUint(splitStr[len(splitStr)-1], 10, 64)
	// sanity check: error should always be nil since regex only allows numbers in last element
	if err != nil {
		panic(fmt.Sprintf("regex allowed non-number value as last split element for chainID: %s", chainID))
	}
	return version
}

// GetSelfHeight is a utility function that returns self height given context
// Version number is retrieved from ctx.ChainID()
func GetSelfHeight(ctx sdk.Context) Height {
	version := ParseChainID(ctx.ChainID())
	return NewHeight(version, uint64(ctx.BlockHeight()))
}
