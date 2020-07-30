package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SenderChainIsSource returns false if the denomination originally came
// from the receiving chain and true otherwise.
func SenderChainIsSource(sourcePort, sourceChannel, denom string) bool {
	// This is the prefix that would have been prefixed to the denomination
	// on sender chain IF and only if the token originally came from the
	// receiving chain.

	voucherPrefix := GetDenomPrefix(sourcePort, sourceChannel)
	return !strings.HasPrefix(denom, voucherPrefix)
}

// ReceiverChainIsSource returns true if the denomination orgininally came
// from the receiving chain and false otherwise.
func ReceiverChainIsSource(prefix, denom string) bool {
	// The prefix passed in should contain the SourcePort and SourceChannel.
	// If  the receiver chain originally sent the token to the sender chain
	// the denom will have the sender's SourcePort and SourceChannel as the
	// prefix.

	return strings.HasPrefix(denom, prefix)
}

// GetDenomPrefix returns the receiving denomination prefix
func GetDenomPrefix(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/", portID, channelID)
}

// GetPrefixedDenom returns the denomination with the portID and channelID prefixed
func GetPrefixedDenom(portID, channelID, baseDenom string) string {
	return fmt.Sprintf("%s/%s/%s", portID, channelID, baseDenom)
}

// GetPrefixedCoin creates a copy of the given coin with the prefixed denom
func GetPrefixedCoin(portID, channelID string, coin sdk.Coin) sdk.Coin {
	return sdk.NewCoin(GetPrefixedDenom(portID, channelID, coin.Denom), coin.Amount)
}

// GetTransferCoin creates a transfer coin with the port ID and channel ID
// prefixed to the base denom.
func GetTransferCoin(portID, channelID, baseDenom string, amount int64) sdk.Coin {
	return sdk.NewInt64Coin(GetPrefixedDenom(portID, channelID, baseDenom), amount)
}
