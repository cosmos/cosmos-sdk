package address


type AddressCdC interface {
	// add encodetextToBytes
	// encodeBytesToText
	encodeTextToBytes(text string) ([]byte, error)
    decodeBytesToText(bytes []byte) (string, error)
}