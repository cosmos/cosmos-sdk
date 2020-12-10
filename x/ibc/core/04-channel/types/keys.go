package types

import (
	"fmt"
	"regexp"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
)

const (
	// SubModuleName defines the IBC channels name
	SubModuleName = "channel"

	// StoreKey is the store key string for IBC channels
	StoreKey = SubModuleName

	// RouterKey is the message route for IBC channels
	RouterKey = SubModuleName

	// QuerierRoute is the querier route for IBC channels
	QuerierRoute = SubModuleName

	// KeyNextChannelSequence is the key used to store the next channel sequence in
	// the keeper.
	KeyNextChannelSequence = "nextChannelSequence"

	// ChannelPrefix is the prefix used when creating a channel identifier
	ChannelPrefix = "channel-"
)

// FormatChannelIdentifier returns the channel identifier with the sequence appended.
// This is a SDK specific format not enforced by IBC protocol.
func FormatChannelIdentifier(sequence uint64) string {
	return fmt.Sprintf("%s%d", ChannelPrefix, sequence)
}

// IsChannelIDFormat checks if a channelID is in the format required on the SDK for
// parsing channel identifiers. The channel identifier must be in the form: `channel-{N}
var IsChannelIDFormat = regexp.MustCompile(`^channel-[0-9]{1,20}$`).MatchString

// IsValidChannelID checks if a channelID is valid and can be parsed to the channel
// identifier format.
func IsValidChannelID(channelID string) bool {
	_, err := ParseChannelSequence(channelID)
	return err == nil
}

// ParseChannelSequence parses the channel sequence from the channel identifier.
func ParseChannelSequence(channelID string) (uint64, error) {
	if !IsChannelIDFormat(channelID) {
		return 0, sdkerrors.Wrap(host.ErrInvalidID, "channel identifier is not in the format: `channel-{N}`")
	}

	sequence, err := host.ParseIdentifier(channelID, ChannelPrefix)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "invalid channel identifier")
	}

	return sequence, nil
}
