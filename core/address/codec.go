package address

// Codec defines an interface to convert addresses from and to string/bytes.
type Codec interface {
	// GetBech32Prefix returns the bech32 prefix
	GetBech32Prefix() string
	// StringToBytes decodes text to bytes
	StringToBytes(text string) ([]byte, error)
	// BytesToString encodes bytes to text
	BytesToString(bz []byte) (string, error)
}
