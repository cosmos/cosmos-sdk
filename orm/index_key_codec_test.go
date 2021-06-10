package orm

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeIndexKey(t *testing.T) {
	specs := map[string]struct {
		srcKey   []byte
		srcRowID RowID
		enc      IndexKeyCodec
		expKey   []byte
		expPanic bool
	}{
		"dynamic length example 1": {
			srcKey:   []byte{0x0, 0x1, 0x2},
			srcRowID: []byte{0x3, 0x4},
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expKey:   []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x2},
		},
		"dynamic length example 2": {
			srcKey:   []byte{0x0, 0x1},
			srcRowID: []byte{0x2, 0x3, 0x4},
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expKey:   []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x3},
		},
		"dynamic length max row ID": {
			srcKey:   []byte{0x0, 0x1},
			srcRowID: []byte(strings.Repeat("a", 255)),
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expKey:   append(append([]byte{0x0, 0x1}, []byte(strings.Repeat("a", 255))...), 0xff),
		},
		"dynamic length panics with empty rowID": {
			srcKey:   []byte{0x0, 0x1},
			srcRowID: []byte{},
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expPanic: true,
		},
		"dynamic length exceeds max row ID": {
			srcKey:   []byte{0x0, 0x1},
			srcRowID: []byte(strings.Repeat("a", 256)),
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expPanic: true,
		},
		"uint64 example": {
			srcKey:   []byte{0x0, 0x1, 0x2},
			srcRowID: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
			enc:      FixLengthIndexKeys(8),
			expKey:   []byte{0x0, 0x1, 0x2, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
		},
		"uint64 panics with empty rowID": {
			srcKey:   []byte{0x0, 0x1},
			srcRowID: []byte{},
			enc:      FixLengthIndexKeys(8),
			expPanic: true,
		},
		"uint64 exceeds max bytes in rowID": {
			srcKey:   []byte{0x0, 0x1},
			srcRowID: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9},
			enc:      FixLengthIndexKeys(8),
			expPanic: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			if spec.expPanic {
				require.Panics(t,
					func() {
						_ = spec.enc.BuildIndexKey(spec.srcKey, spec.srcRowID)
					})
				return
			}
			got := spec.enc.BuildIndexKey(spec.srcKey, spec.srcRowID)
			assert.Equal(t, spec.expKey, got)
		})
	}
}
func TestDecodeIndexKey(t *testing.T) {
	specs := map[string]struct {
		srcKey   []byte
		enc      IndexKeyCodec
		expRowID RowID
	}{
		"dynamic length example 1": {
			srcKey:   []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x2},
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expRowID: []byte{0x3, 0x4},
		},
		"dynamic length example 2": {
			srcKey:   []byte{0x0, 0x1, 0x2, 0x3, 0x4, 0x3},
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expRowID: []byte{0x2, 0x3, 0x4},
		},
		"dynamic length max row ID": {
			srcKey:   append(append([]byte{0x0, 0x1}, []byte(strings.Repeat("a", 255))...), 0xff),
			enc:      Max255DynamicLengthIndexKeyCodec{},
			expRowID: []byte(strings.Repeat("a", 255)),
		},
		"uint64 example": {
			srcKey:   []byte{0x0, 0x1, 0x2, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
			expRowID: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8},
			enc:      FixLengthIndexKeys(8),
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			gotRow := spec.enc.StripRowID(spec.srcKey)
			assert.Equal(t, spec.expRowID, gotRow)
		})
	}
}
