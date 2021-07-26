package v044

var (
	DenomAddressPrefix = []byte{0x03}
)

func CreateAddressDenomPrefix(denom string) []byte {
	key := append(DenomAddressPrefix, []byte(denom)...)
	return append(key, 0)
}
