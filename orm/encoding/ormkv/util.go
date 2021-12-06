package ormkv

import (
	"bytes"
	"io"
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
