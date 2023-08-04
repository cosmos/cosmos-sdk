package v2

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

const (
	// ModuleName is the name of the module
	ModuleName = "staking"
)

var (
	ValidatorsByConsAddrKey = []byte{0x22} // prefix for validators by consensus address
	HistoricalInfoKey       = []byte{0x50} // prefix for the historical info
)

// GetHistoricalInfoKey returns a key prefix for indexing HistoricalInfo objects.
func GetHistoricalInfoKey(height int64) []byte {
	return append(HistoricalInfoKey, []byte(strconv.FormatInt(height, 10))...)
}

// GetValidatorByConsAddrKey creates the key for the validator with pubkey
// VALUE: validator operator address ([]byte)
func GetValidatorByConsAddrKey(addr sdk.ConsAddress) []byte {
	return append(ValidatorsByConsAddrKey, address.MustLengthPrefix(addr)...)
}
