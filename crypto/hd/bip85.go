package hd

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// DeriveBIP85Entropy derives numBytes (1–64) of deterministic entropy from a
// BIP32 master private key and chain code, as specified in BIP85. The absolute
// path must start with m/83696968' and contain only hardened components.
// Applications must use distinct paths for distinct algorithms and accounts.
// The returned entropy is secret; callers must not log it or expose the BIP32
// private keys used to derive it. Invalid BIP32 keys fail rather than silently
// selecting a different path.
func DeriveBIP85Entropy(masterPriv, chainCode [32]byte, path string, numBytes int) ([]byte, error) {
	defer clear(masterPriv[:])
	defer clear(chainCode[:])
	if numBytes < 1 || numBytes > sha512.Size {
		return nil, fmt.Errorf("BIP85 entropy length must be between 1 and 64 bytes")
	}
	parts := strings.Split(path, "/")
	if len(parts) < 4 || len(parts) > 256 || parts[0] != "m" || parts[1] != "83696968'" {
		return nil, fmt.Errorf("BIP85 path must be m/83696968'/application'/index' with at most 255 components")
	}
	indices := make([]uint32, len(parts)-1)
	for i, part := range parts[1:] {
		if !strings.HasSuffix(part, "'") {
			return nil, fmt.Errorf("BIP85 path component %d must be hardened", i+1)
		}
		index, err := strconv.ParseUint(strings.TrimSuffix(part, "'"), 10, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid BIP85 path component %d: %w", i+1, err)
		}
		indices[i] = uint32(index) | 0x80000000
	}

	var key, offset secp.ModNScalar
	defer key.Zero()
	defer offset.Zero()
	if key.SetBytes(&masterPriv) != 0 || key.IsZero() {
		return nil, fmt.Errorf("invalid BIP32 master private key")
	}
	var data [37]byte // 0x00 || ser256(k) || ser32(i), hardened CKDpriv
	defer clear(data[:])
	for i, index := range indices {
		copy(data[1:33], masterPriv[:])
		binary.BigEndian.PutUint32(data[33:], index)
		mac := hmac.New(sha512.New, chainCode[:])
		_, _ = mac.Write(data[:])
		digest := mac.Sum(nil)
		overflow := offset.SetByteSlice(digest[:32])
		key.Add(&offset)
		copy(chainCode[:], digest[32:])
		clear(digest)
		if overflow || key.IsZero() {
			return nil, fmt.Errorf("invalid BIP32 child at component %d; choose another path", i+1)
		}
		masterPriv = key.Bytes()
	}

	mac := hmac.New(sha512.New, []byte("bip-entropy-from-k"))
	_, _ = mac.Write(masterPriv[:])
	digest := mac.Sum(nil)
	entropy := make([]byte, numBytes)
	copy(entropy, digest)
	clear(digest)
	return entropy, nil
}
