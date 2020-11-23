package host

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

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
