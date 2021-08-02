package address

type AddressCodec interface {
	// AddressStringToBytes encodes text to bytes
	AddressStringToBytes(text string) ([]byte, error)
	// AddressBytesToString decodes bytes to text
	AddressBytesToString(bytes []byte) (string, error)
}
