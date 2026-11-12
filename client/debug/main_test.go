package debug

import (
	"io"
	"os"
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

			// the command writes its result with fmt.Printf, so capture os.Stdout
			r, w, err := os.Pipe()
			require.NoError(t, err)
			stdout := os.Stdout
			os.Stdout = w
			err = cmd.Execute()
			os.Stdout = stdout
			require.NoError(t, w.Close())
			out, readErr := io.ReadAll(r)
			require.NoError(t, readErr)

			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, string(out))
		})
	}
}
