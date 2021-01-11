package host

import (
	"strconv"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// ParseIdentifier parses the sequence from the identifier using the provided prefix. This function
// does not need to be used by counterparty chains. SDK generated connection and channel identifiers
// are required to use this format.
func ParseIdentifier(identifier, prefix string) (uint64, error) {
	if !strings.HasPrefix(identifier, prefix) {
		return 0, sdkerrors.Wrapf(ErrInvalidID, "identifier doesn't contain prefix `%s`", prefix)
	}

	splitStr := strings.Split(identifier, prefix)
	if len(splitStr) != 2 {
		return 0, sdkerrors.Wrapf(ErrInvalidID, "identifier must be in format: `%s{N}`", prefix)
	}

	// sanity check
	if splitStr[0] != "" {
		return 0, sdkerrors.Wrapf(ErrInvalidID, "identifier must begin with prefix %s", prefix)
	}

	sequence, err := strconv.ParseUint(splitStr[1], 10, 64)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "failed to parse identifier sequence")
	}
	return sequence, nil
}

// ParseConnectionPath returns the connection ID from a full path. It returns
// an error if the provided path is invalid.
func ParseConnectionPath(path string) (string, error) {
	split := strings.Split(path, "/")
	if len(split) != 2 {
		return "", sdkerrors.Wrapf(ErrInvalidPath, "cannot parse connection path %s", path)
	}

	return split[1], nil
}

// ParseChannelPath returns the port and channel ID from a full path. It returns
// an error if the provided path is invalid.
func ParseChannelPath(path string) (string, string, error) {
	split := strings.Split(path, "/")
	if len(split) < 5 {
		return "", "", sdkerrors.Wrapf(ErrInvalidPath, "cannot parse channel path %s", path)
	}

	if split[1] != KeyPortPrefix || split[3] != KeyChannelPrefix {
		return "", "", sdkerrors.Wrapf(ErrInvalidPath, "cannot parse channel path %s", path)
	}

	return split[2], split[4], nil
}

// MustParseConnectionPath returns the connection ID from a full path. Panics
// if the provided path is invalid.
func MustParseConnectionPath(path string) string {
	connectionID, err := ParseConnectionPath(path)
	if err != nil {
		panic(err)
	}
	return connectionID
}

// MustParseChannelPath returns the port and channel ID from a full path. Panics
// if the provided path is invalid.
func MustParseChannelPath(path string) (string, string) {
	portID, channelID, err := ParseChannelPath(path)
	if err != nil {
		panic(err)
	}
	return portID, channelID
}
