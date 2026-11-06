package cosmovisor

import (
	"testing"

	dbm "github.com/cometbft/cometbft-db"
	"github.com/stretchr/testify/require"
)

func TestParseDBBackend(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		expect    dbm.BackendType
		expectErr bool
	}{
		{name: "quoted with newline", input: "\"goleveldb\"\n", expect: "goleveldb"},
		{name: "quoted", input: "\"goleveldb\"", expect: "goleveldb"},
		{name: "literal string", input: "'pebbledb'\n", expect: "pebbledb"},
		{name: "bare", input: "goleveldb\n", expect: "goleveldb"},
		{name: "surrounding whitespace", input: "  \"rocksdb\"  \r\n", expect: "rocksdb"},
		{name: "unknown backend is left to NewDB", input: "future_backend_2", expect: "future_backend_2"},
		{name: "empty", input: "", expectErr: true},
		{name: "only newline", input: "\n", expectErr: true},
		{name: "empty quoted", input: "\"\"\n", expectErr: true},
		{name: "space in value", input: "\"go leveldb\"", expectErr: true},
		{name: "command error text", input: "Error: unknown key\n", expectErr: true},
		{name: "extra output line", input: "\"goleveldb\"\nWARN extra line\n", expectErr: true},
		{name: "unterminated quote", input: "\"goleveldb", expectErr: true},
		{name: "escaped newline", input: "\"go\\nleveldb\"", expectErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseDBBackend([]byte(tc.input))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, got)
		})
	}
}
