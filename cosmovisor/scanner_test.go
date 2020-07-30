package cosmovisor

import (
	"bufio"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitForInfo(t *testing.T) {
	cases := map[string]struct {
		write         []string
		expectUpgrade *UpgradeInfo
		expectErr     bool
	}{
		"no match": {
			write: []string{"some", "random\ninfo\n"},
		},
		"match height with no info": {
			write: []string{"first line\n", `UPGRADE "myname" NEEDED at height: 123: `, "\nnext line\n"},
			expectUpgrade: &UpgradeInfo{
				Name:   "myname",
				Height: 123,
				Info:   "",
			},
		},
		"match height with info": {
			write: []string{"first line\n", `UPGRADE "take2" NEEDED at height: 123:   DownloadData here!`, "\nnext line\n"},
			expectUpgrade: &UpgradeInfo{
				Name:   "take2",
				Height: 123,
				Info:   "DownloadData",
			},
		},
		"match time with no info": {
			write: []string{"first line\n", `UPGRADE "timer" NEEDED at time: 2020-04-01T11:22:33Z:   `, "\nnext line\n"},
			expectUpgrade: &UpgradeInfo{
				Name: "timer",
				Time: "2020-04-01T11:22:33Z",
				Info: "",
			},
		},
		"match time with info": {
			write: []string{"first line\n", `UPGRADE "april" NEEDED at time: 2020-04-01T11:22:33Z: https://april.foo.rs/hahaha  `, "\nnext line\n"},
			expectUpgrade: &UpgradeInfo{
				Name: "april",
				Time: "2020-04-01T11:22:33Z",
				Info: "https://april.foo.rs/hahaha",
			},
		},
		"chunks": {
			write: []string{"first l", "ine\nERROR 2020-02-03T11:22:33Z: UPGRADE ", `"split" NEEDED at height: `, "789:   {\"foo\":123} asgsdg", "  \n LOG: next line"},
			expectUpgrade: &UpgradeInfo{
				Name:   "split",
				Height: 789,
				Info:   `{"foo":123}`,
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
					assert.NoError(t, err)
					assert.Equal(t, len(line), n)
				}
				w.Close()
			}()

			// now scan the info
			info, err := WaitForUpdate(scan)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectUpgrade, info)
		})
	}
}
