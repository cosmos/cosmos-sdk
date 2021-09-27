package v044

var (
	DenomAddressPrefix = []byte{0x03}
)

func CreateAddressDenomPrefix(denom string) []byte {
	key := make([]byte, len(DenomAddressPrefix)+len(denom)+1)
	copy(key, DenomAddressPrefix)
	copy(key[len(DenomAddressPrefix):], denom)
	// key[last] = 0 - not needed, because this is the default
	return key
}
