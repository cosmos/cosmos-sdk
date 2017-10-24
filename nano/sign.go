package nano

import (
	"bytes"
	"crypto/sha512"

	"github.com/pkg/errors"
)

const (
	App       = 0x80
	Init      = 0x00
	Update    = 0x01
	Digest    = 0x02
	MaxChunk  = 253
	KeyLength = 32
	SigLength = 64
)

var separator = []byte{0, 0xCA, 0xFE, 0}

func generateSignRequests(payload []byte) [][]byte {
	// nice one-shot
	digest := []byte{App, Digest}
	if len(payload) < MaxChunk {
		return [][]byte{append(digest, payload...)}
	}

	// large payload is multi-chunk
	result := [][]byte{{App, Init}}
	update := []byte{App, Update}
	for len(payload) > MaxChunk {
		msg := append(update, payload[:MaxChunk]...)
		payload = payload[MaxChunk:]
		result = append(result, msg)
	}
	result = append(result, append(update, payload...))
	result = append(result, digest)
	return result
}

func parseDigest(resp []byte) (key, sig []byte, err error) {
	if resp[0] != App || resp[1] != Digest {
		return nil, nil, errors.New("Invalid header")
	}
	resp = resp[2:]
	if len(resp) != KeyLength+SigLength+len(separator) {
		return nil, nil, errors.Errorf("Incorrect length: %d", len(resp))
	}

	key, resp = resp[:KeyLength], resp[KeyLength:]
	if !bytes.Equal(separator, resp[:len(separator)]) {
		return nil, nil, errors.New("Cannot find 0xCAFE")
	}

	sig = resp[len(separator):]
	return key, sig, nil
}

func hashMsg(data []byte) []byte {
	res := sha512.Sum512(data)
	return res[:]
}
