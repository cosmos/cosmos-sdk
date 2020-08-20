package host

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// RemovePath is an util function to remove a path from a set.
func RemovePath(paths []string, path string) ([]string, bool) {
	for i, p := range paths {
		if p == path {
			return append(paths[:i], paths[i+1:]...), true
		}
	}
	return paths, false
}

// ParseClientPath returns the client ID from a full path. It returns an error if the provided path
// is not in the format "clients/{clientID}..." or if the client identifier is invalid.
func ParseClientPath(path string) (string, error) {
	split := strings.Split(path, "/")
	if len(split) < 2 || split[0] != "clients" {
		return "", sdkerrors.Wrapf(ErrInvalidPath, "cannot parse client path %s", path)
	}

	clientdID := split[1]
	if err := ClientIdentifierValidator(clientdID); err != nil {
		return "", err
	}

	return clientdID, nil
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

	if split[1] != "ports" || split[3] != "channels" {
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
