package types

import (
	"crypto/sha256"
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateID creates a deterministic pseudorandom integer using data in the sdk context as entropy sources.
// The top bit of this int is always set to 1 to avoid conflicts with legacy IDs which were generated
// via incrementing a sequence number.
func GenerateID(ctx sdk.Context) uint64 {
	blkHeader := ctx.BlockHeader()

	h := sha256.New()
	_ = binary.Write(h, binary.BigEndian, blkHeader.Height)
	h.Write(blkHeader.AppHash)
	_ = binary.Write(h, binary.BigEndian, int64(ctx.TxIndex()))
	_ = binary.Write(h, binary.BigEndian, int64(ctx.MsgIndex()))

	digest := h.Sum(nil)
	x := binary.BigEndian.Uint64(digest[:8])
	x |= uint64(1) << 63 // force top bit 1
	return x
}
