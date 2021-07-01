package v044

var (
	DenomPrefix = []byte{0x03}
)

func CreateDenomPrefix(denom string) []byte {
	key := append(DenomPrefix, []byte(denom)...)
	return append(key, 0)
}
