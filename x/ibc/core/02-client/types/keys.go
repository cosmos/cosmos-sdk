package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

const (
	// SubModuleName defines the IBC client name
	SubModuleName string = "client"

	// RouterKey is the message route for IBC client
	RouterKey string = SubModuleName

	// QuerierRoute is the querier route for IBC client
	QuerierRoute string = SubModuleName

	// KeyNextClientSequence is the key used to store the next client sequence in
	// the keeper.
	KeyNextClientSequence = "nextClientSequence"
)

// FormatClientIdentifier returns the client identifier with the sequence appended.
// This is a SDK specific format not enforced by IBC protocol.
func FormatClientIdentifier(clientType string, sequence uint64) string {
	return fmt.Sprintf("%s-%d", clientType, sequence)
}

// IsClientIDFormat checks if a clientID is in the format required on the SDK for
// parsing client identifiers. The client identifier must be in the form: `{client-type}-{N}
var IsClientIDFormat = regexp.MustCompile(`^.*[^-]-[0-9]{1,20}$`).MatchString

// IsValidClientID checks if the clientID is valid and can be parsed into the client
// identifier format.
func IsValidClientID(clientID string) bool {
	_, _, err := ParseClientIdentifier(clientID)
	return err == nil
}

// ParseClientIdentifier parses the client type and sequence from the client identifier.
func ParseClientIdentifier(clientID string) (string, uint64, error) {
	if !IsClientIDFormat(clientID) {
		return "", 0, sdkerrors.Wrapf(host.ErrInvalidID, "invalid client identifier %s is not in format: `{client-type}-{N}`", clientID)
	}

	splitStr := strings.Split(clientID, "-")
	lastIndex := len(splitStr) - 1

	clientType := strings.Join(splitStr[:lastIndex], "-")
	if strings.TrimSpace(clientType) == "" {
		return "", 0, sdkerrors.Wrap(host.ErrInvalidID, "client identifier must be in format: `{client-type}-{N}` and client type cannot be blank")
	}

	sequence, err := strconv.ParseUint(splitStr[lastIndex], 10, 64)
	if err != nil {
		return "", 0, sdkerrors.Wrap(err, "failed to parse client identifier sequence")
	}

	return clientType, sequence, nil
}
