package address

// Codec defines the interface for encoding and decoding addresses
type Codec interface {
	// StringToBytes decodes text to bytes
	StringToBytes(text string) ([]byte, error)
	// BytesToString encodes bytes to text
	BytesToString(bz []byte) (string, error)
}
