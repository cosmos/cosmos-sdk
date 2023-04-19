package address

// Codec defines an interface to convert addresses from and to string/bytes.
type Codec interface {
	// StringToBytes decodes a string address (in bech32 format for example) to
	// its binary representation. An empty string returns nil and not an error.
	StringToBytes(text string) ([]byte, error)

	// BytesToString encodes a binary address to its string representation (bech32
	// for example). An empty byte array returns an empty string and not an error.
	BytesToString(bz []byte) (string, error)
}
