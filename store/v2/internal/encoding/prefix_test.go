package encoding

import (
	"bytes"
	"testing"
)

func TestBuildPrefixWithVersion(t *testing.T) {
	testcases := []struct {
		prefix  string
		version uint64
		want    []byte
	}{
		{"", 0, []byte{0, 0, 0, 0, 0, 0, 0, 0, '/'}},
		{"a", 0, []byte{'a', 0, 0, 0, 0, 0, 0, 0, 0, '/'}},
		{"b", 1, []byte{'b', 0, 0, 0, 0, 0, 0, 0, 1, '/'}},
		{"s/k/removed/", 2, []byte{'s', '/', 'k', '/', 'r', 'e', 'm', 'o', 'v', 'e', 'd', '/', 0, 0, 0, 0, 0, 0, 0, 2, '/'}},
		{"s/k/", 1234567890, []byte{'s', '/', 'k', '/', 0, 0, 0, 0, 73, 150, 2, 210, '/'}},
	}

	for _, tc := range testcases {
		got := BuildPrefixWithVersion(tc.prefix, tc.version)
		if !bytes.Equal(got, tc.want) {
			t.Fatalf("BuildPrefixWithVersion(%q, %d) = %v, want %v", tc.prefix, tc.version, got, tc.want)
		}
	}
}
