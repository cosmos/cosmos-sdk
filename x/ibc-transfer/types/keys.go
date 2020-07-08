package types

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the IBC transfer name
	ModuleName = "transfer"

	// Version defines the current version the IBC tranfer
	// module supports
	Version = "ics20-1"

	// PortID is the default port id that transfer module binds to
	PortID = "transfer"

	// StoreKey is the store key string for IBC transfer
	StoreKey = ModuleName

	// RouterKey is the message route for IBC transfer
	RouterKey = ModuleName

	// QuerierRoute is the querier route for IBC transfer
	QuerierRoute = ModuleName
)

// PortKey defines the key to store the port ID in store
var PortKey = []byte{0x01}

// GetEscrowAddress returns the escrow address for the specified channel
//
// CONTRACT: this assumes that there's only one bank bridge module that owns the
// port associated with the channel ID so that the address created is actually
// unique.
func GetEscrowAddress(portID, channelID string) sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(portID + channelID)))
}

// GetDenomPrefix returns the receiving denomination prefix
func GetDenomPrefix(portID, channelID string) string {
	return fmt.Sprintf("%s/%s/", portID, channelID)
}

// GetPrefixedCoins creates a copy of the given coins with the denom updated with the prefix.
func GetPrefixedCoins(portID, channelID string, coins ...sdk.Coin) sdk.Coins {
	prefixedCoins := make(sdk.Coins, len(coins))
	for i := range coins {
		prefixedCoins[i] = sdk.NewCoin(GetDenomPrefix(portID, channelID)+coins[i].Denom, coins[i].Amount)
	}
	return prefixedCoins
}
