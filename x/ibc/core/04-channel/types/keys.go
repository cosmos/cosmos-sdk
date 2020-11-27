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

// IsValidChannelID checks if a channelID is in the format required for parsing channel
// identifier. The channel identifier must be in the form: `connection-{N}
var IsValidChannelID = regexp.MustCompile(`^channel-[0-9]{1,20}$`).MatchString

// FormatChannelIdentifier returns the channel identifier with the sequence appended.
func FormatChannelIdentifier(sequence uint64) string {
	return fmt.Sprintf("%s%d", ChannelPrefix, sequence)
}

// ParseChannelSequence parses the channel sequence from the channel identifier.
func ParseChannelSequence(channelID string) (uint64, error) {
	sequence, err := host.ParseIdentifier(channelID, ChannelPrefix)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "invalid channel identifier")
	}

	return sequence, nil
}
