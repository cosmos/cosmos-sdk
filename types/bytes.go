package types

import "encoding/hex"

// String encode bytes using hex encoder
func (m *PBBytes) String() string {
	return hex.EncodeToString(m.Bytes)
}
