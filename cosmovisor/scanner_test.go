package cosmovisor_test

import (
	"bufio"
	"io"
	"testing"

	"github.com/cosmos/cosmos-sdk/cosmovisor"

	"github.com/stretchr/testify/require"
)

func TestWaitForInfo(t *testing.T) {
	cases := map[string]struct {
		write         []string
		expectUpgrade *cosmovisor.UpgradeInfo
		expectErr     bool
	}{
		"no match": {
			write: []string{"some", "random\ninfo\n"},
		},
		"match name with no info": {
			write: []string{"first line\n", `UPGRADE "myname" NEEDED at height: 123: `, "\nnext line\n"},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "myname",
				Info: "",
			},
		},
		"match name with info": {
			write: []string{"first line\n", `UPGRADE "take2" NEEDED at height: 123:   DownloadData here!`, "\nnext line\n"},
			expectUpgrade: &cosmovisor.UpgradeInfo{
				Name: "take2",
				Info: "DownloadData",
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r, w := io.Pipe()
			scan := bufio.NewScanner(r)

			// write all info in separate routine
			go func() {
				for _, line := range tc.write {
					n, err := w.Write([]byte(line))
					require.NoError(t, err)
					require.Equal(t, len(line), n)
				}
				w.Close()
			}()

			// now scan the info
			info, err := cosmovisor.WaitForUpdate(scan)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectUpgrade, info)
		})
	}
}
