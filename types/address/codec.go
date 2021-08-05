package address

type Codec interface {
	// ConvertAddressStringToBytes decodes text to bytes
	ConvertAddressStringToBytes(text string) ([]byte, error)
	// ConvertAddressBytesToString encodes bytes to text
	ConvertAddressBytesToString(bz []byte) (string, error)
}

func Ver
