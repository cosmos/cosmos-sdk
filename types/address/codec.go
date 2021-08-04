package address

type Codec interface {
	// addressStringToBytes encodes text to bytes
	ConvertAddressStringToBytes(text string) ([]byte, error)
	// addressBytesToString encodes bytes to text
	ConvertAddressBytesToString(bz []byte) (string, error)
}
