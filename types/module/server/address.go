package server

import "crypto/sha256"

func AddressHash(prefix string, contents []byte) []byte {
	preImage := []byte(prefix)
	if len(contents) != 0 {
		preImage = append(preImage, 0)
		preImage = append(preImage, contents...)
	}
	sum := sha256.Sum256(preImage)
	return sum[:20]
}
