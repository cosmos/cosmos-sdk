package encodeutil

import (
	"bytes"
	"encoding/binary"
	"io"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// SkipPrefix skips the provided prefix in the reader or returns an error.
// This is used for efficient logical decoding of keys.
func SkipPrefix(r *bytes.Reader, prefix []byte) error {
	n := len(prefix)
	if n > 0 {
		// we skip checking the prefix for performance reasons because we assume
		// that it was checked by the caller
		_, err := r.Seek(int64(n), io.SeekCurrent)
		return err
	}
	return nil
}

// AppendVarUInt32 creates a new key prefix, by encoding and appending a
// var-uint32 to the provided prefix.
func AppendVarUInt32(prefix []byte, x uint32) []byte {
	prefixLen := len(prefix)
	res := make([]byte, prefixLen+binary.MaxVarintLen32)
	copy(res, prefix)
	n := binary.PutUvarint(res[prefixLen:], uint64(x))
	return res[:prefixLen+n]
}

// ValuesOf takes the arguments and converts them to protoreflect.Value's.
func ValuesOf(values ...interface{}) []protoreflect.Value {
	n := len(values)
	res := make([]protoreflect.Value, n)
	for i := 0; i < n; i++ {
		res[i] = protoreflect.ValueOf(values[i])
	}
	return res
}
