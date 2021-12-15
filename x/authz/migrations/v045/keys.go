package v045

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Keys for store prefixes
// Items are stored with the following key: values
//
// - 0x02<grant_expiration_Bytes>: GGMTriple
//
var (
	GrantQueuePrefix = []byte{0x02}
)

// GrantQueueKey - return grant queue store key
// Key format is
//
// - 0x02<grant_expiration_Bytes>
func GrantQueueKey(expiration time.Time) []byte {
	return append(GrantQueuePrefix, sdk.FormatTimeBytes(expiration)...)
}
