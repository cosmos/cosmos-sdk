package types

import (
	"fmt"
	"regexp"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

const (
	// SubModuleName defines the IBC connection name
	SubModuleName = "connection"

	// StoreKey is the store key string for IBC connections
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC connections
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC connections
	QuerierRoute = SubModuleName

	// KeyNextConnectionSequence is the key used to store the next connection sequence in
	// the keeper.
	KeyNextConnectionSequence = "nextConnectionSequence"

	// ConnectionPrefix is the prefix used when creating a connection identifier
	ConnectionPrefix = "connection-"
)

// IsValidConnectionID checks if a connectionID is in the format required for parsing client
// identifier. The client identifier must be in the form: `connection-{N}
var IsValidConnectionID = regexp.MustCompile(`^connection-[0-9]{1,20}$`).MatchString

// FormatConnectionIdentifier returns the connection identifier with the sequence appended.
func FormatConnectionIdentifier(sequence uint64) string {
	return fmt.Sprintf("%s%d", ConnectionPrefix, sequence)
}

// ParseConnectionSequence parses the connection sequence from the connection identifier.
func ParseConnectionSequence(connectionID string) (uint64, error) {
	sequence, err := host.ParseIdentifier(connectionID, ConnectionPrefix)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "invalid connection identifier")
	}

	return sequence, nil
}
