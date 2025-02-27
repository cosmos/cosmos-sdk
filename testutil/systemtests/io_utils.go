package systemtests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// MustCopyFile copies the file from the source path `src` to the destination path `dest` and returns an open file handle to `dest`.
func MustCopyFile(src, dest string) *os.File {
	in, err := os.Open(src)
	if err != nil {
		panic(fmt.Sprintf("failed to open %q: %v", src, err))
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		panic(fmt.Sprintf("failed to create %q: %v", dest, err))
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		panic(fmt.Sprintf("failed to copy from %q to %q: %v", src, dest, err))
	}
	return out
}

// MustCopyFilesInDir copies all files (excluding directories) from the source directory `src` to the destination directory `dest`.
func MustCopyFilesInDir(src, dest string) {
	err := os.MkdirAll(dest, 0o750)
	if err != nil {
		panic(fmt.Sprintf("failed to create %q: %v", dest, err))
	}

	fs, err := os.ReadDir(src)
	if err != nil {
		panic(fmt.Sprintf("failed to read dir %q: %v", src, err))
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}
		_ = MustCopyFile(filepath.Join(src, f.Name()), filepath.Join(dest, f.Name()))
	}
}

// StoreTempFile creates a temporary file in the test's temporary directory with the provided content.
// It returns a pointer to the created file. Errors during the process are handled with test assertions.
func StoreTempFile(t *testing.T, content []byte) *os.File {
	t.Helper()
	out, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)
	_, err = io.Copy(out, bytes.NewReader(content))
	require.NoError(t, err)
	require.NoError(t, out.Close())
	return out
}
