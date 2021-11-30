package ormkv

import (
	"bytes"
	"io"
)

func skipPrefix(r *bytes.Reader, prefix []byte) error {
	n := len(prefix)
	if n > 0 {
		// we skip checking the prefix for performance reasons because we assume
		// that it was checked by the caller
		_, err := r.Seek(int64(n), io.SeekCurrent)
		return err
	}
	return nil
}
