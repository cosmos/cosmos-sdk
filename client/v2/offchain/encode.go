package offchain

import "encoding/base64"

const (
	noEncoder  = "no-encoding"
	b64Encoder = "base64"
)

type encodingFunc = func([]byte) (string, error)

func noEncoding(digest []byte) (string, error) {
	return string(digest), nil
}

func base64Encoding(digest []byte) (string, error) {
	return base64.StdEncoding.EncodeToString(digest), nil
}

func getEncoder(encoder string) encodingFunc {
	switch encoder {
	case noEncoder:
		return noEncoding
	case b64Encoder:
		return base64Encoding
	}
	return noEncoding
}
