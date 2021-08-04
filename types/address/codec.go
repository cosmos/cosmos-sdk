package address

type Codec interface {
	// ConvertAddressStringToBytes encodes text to bytes
	ConvertAddressStringToBytes(text string) ([]byte, error)
	// ConvertAddressBytesToString encodes bytes to text
	ConvertAddressBytesToString(bz []byte) (string, error)
}
