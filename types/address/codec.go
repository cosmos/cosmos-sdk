package address

type Codec interface {
	// AddressStringToBytes decodes text to bytes
	StringToBytes(text string) ([]byte, error)
	// BytesToString encodes bytes to text
	BytesToString(bz []byte) (string, error)
}
