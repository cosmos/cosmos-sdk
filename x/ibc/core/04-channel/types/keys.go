package types

import (
	"fmt"
	"strconv"
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
func FormatChannelIdentifier(sequence uint64) string {
	return fmt.Sprintf("%s%d", ChannelPrefix, sequence)
}

// IsValidChannelID return true if the channel identifier is valid.
func IsValidChannelID(channelID string) bool {
	_, err := ParseChannelSequence(channelID)
	return err == nil
}

// ParseChannelSequence parses the channel sequence from the channel identifier.
func ParseChannelSequence(channelID string) (uint64, error) {
	if !strings.HasPrefix(channelID, ChannelPrefix) {
		return 0, sdkerrors.Wrapf(ErrInvalidChannelIdentifier, "doesn't contain prefix `%s`", ChannelPrefix)
	}

	splitStr := strings.Split(channelID, ChannelPrefix)
	if len(splitStr) != 2 {
		return 0, sdkerrors.Wrap(ErrInvalidChannelIdentifier, "channel identifier must be in format: `channel-{N}`")
	}

	sequence, err := strconv.ParseUint(splitStr[1], 10, 64)
	if err != nil {
		return 0, sdkerrors.Wrap(err, "failed to parse channel identifier sequence")
	}
	return sequence, nil
}
