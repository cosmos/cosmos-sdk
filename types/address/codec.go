package address

type AddressCodec interface {
	// AddressStringToBytes encodes text to bytes
	AddressStringToBytes(text string) ([]byte, error)
	// AddressBytesToString encodes bytes to text
	AddressBytesToString(bz []byte) (string, error)
}
