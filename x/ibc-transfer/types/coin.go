package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
