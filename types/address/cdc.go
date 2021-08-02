package address


type AddressCdC interface {
	// EncodeTextToBytes encodes text to bytes
	EncodeTextToBytes(text string) ([]byte, error)
	// DecodeBytesToText decodes bytes to text
    DecodeBytesToText(bytes []byte) (string, error)
}