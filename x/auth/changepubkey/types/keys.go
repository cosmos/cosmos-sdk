package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// constants
var (
	// module name
	ModuleName = "changepubkey"

	// StoreKey is string representation of the store key for changepubkey
	StoreKey = ModuleName

	// QuerierRoute is the querier route for changepubkey
	QuerierRoute = ModuleName

	// AttributeValueCategory is an alias for the message event value.
	AttributeValueCategory = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// KeyPrefixPubKeyHistory defines history of PubKey history of an account
	KeyPrefixPubKeyHistory = []byte{0x01} // prefix for the timestamps in pubkey history queue
)

// GetPubKeyHistoryKey returns the prefix key used for getting a set of history
// where pubkey endTime is after a specific time
func GetPubKeyHistoryKey(timestamp time.Time) []byte {
	timeBz := sdk.FormatTimeBytes(timestamp)
	timeBzL := len(timeBz)
	prefixL := len(KeyPrefixPubKeyHistory)

	bz := make([]byte, prefixL+8+timeBzL)

	// copy the prefix
	copy(bz[:prefixL], KeyPrefixPubKeyHistory)

	// copy the encoded time bytes length
	copy(bz[prefixL:prefixL+8], sdk.Uint64ToBigEndian(uint64(timeBzL)))

	// copy the encoded time bytes
	copy(bz[prefixL+8:prefixL+8+timeBzL], timeBz)
	return bz
}
