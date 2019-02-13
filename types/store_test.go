package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrefixEndBytes(t *testing.T) {
	var testCases = []struct {
		prefix   []byte
		expected []byte
	}{
		{[]byte{byte(55), byte(255), byte(255), byte(0)}, []byte{byte(55), byte(255), byte(255), byte(1)}},
		{[]byte{byte(55), byte(255), byte(255), byte(15)}, []byte{byte(55), byte(255), byte(255), byte(16)}},
		{[]byte{byte(55), byte(200), byte(255)}, []byte{byte(55), byte(201)}},
		{[]byte{byte(55), byte(255), byte(255)}, []byte{byte(56)}},
		{[]byte{byte(255), byte(255), byte(255)}, nil},
		{[]byte{byte(255)}, nil},
		{nil, nil},
	}

	for _, test := range testCases {
		end := PrefixEndBytes(test.prefix)
		require.Equal(t, test.expected, end)
	}
}

func TestCommitID(t *testing.T) {
	var empty CommitID
	require.True(t, empty.IsZero())

	var nonempty = CommitID{
		Version: 1,
		Hash:    []byte("testhash"),
	}
	require.False(t, nonempty.IsZero())
}
