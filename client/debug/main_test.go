package debug

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRawBytes(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []byte
		wantErr bool
	}{
		{
			name:  "accepts unsigned byte values",
			input: "[10 21 13 255]",
			want:  []byte{10, 21, 13, 255},
		},
		{
			name:  "accepts flexible whitespace",
			input: "[72\t101\n108 108 111]",
			want:  []byte("Hello"),
		},
		{
			name:    "rejects negative values",
			input:   "[-1]",
			wantErr: true,
		},
		{
			name:    "rejects values above byte range",
			input:   "[256]",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseRawBytes(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
