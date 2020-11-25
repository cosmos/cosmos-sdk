package keeper

// Keys for supply store
// Items are stored with the following key: values
//
// - 0x00: Supply
var (
	SupplyKey = []byte{0x00}
)

// getTokenSupplyKey gets the store key of a supply for a token
func getTokenSupplyKey(denom string) []byte {
	return append(SupplyKey, []byte(denom)...)
}
