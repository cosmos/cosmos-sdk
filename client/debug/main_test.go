package debug

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawBytesCmd(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"example from help", "[72 101 108 108 111]", "48656C6C6F\n", false},
		{"bytes above 127", "[10 21 13 255]", "0A150DFF\n", false},
		{"value above 255", "[256]", "", true},
		{"negative value", "[-1]", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := RawBytesCmd()
			cmd.SetArgs([]string{tc.input})
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			var out bytes.Buffer
			cmd.SetOut(&out)
			err := cmd.Execute()

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, out.String())
		})
	}
}
