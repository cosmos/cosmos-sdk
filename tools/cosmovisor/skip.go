package cosmovisor

import (
	"fmt"
	"strconv"
	"strings"

	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// IsSkipUpgradeHeight checks if pre-upgrade script must be run.
// If the height in the upgrade plan matches any of the heights provided in --unsafe-skip-upgrades, the script is not run.
func IsSkipUpgradeHeight(args []string, upgradeInfo upgradetypes.Plan) bool {
	skipUpgradeHeights := UpgradeSkipHeights(args)
	for _, h := range skipUpgradeHeights {
		if h == int(upgradeInfo.Height) {
			return true
		}
	}
	return false
}

// UpgradeSkipHeights gets all the heights provided when
// simd start --unsafe-skip-upgrades <height1> <optional_height_2> ... <optional_height_N>
func UpgradeSkipHeights(args []string) []int {
	var heights []int
	for i, arg := range args {
		if arg == fmt.Sprintf("--%s", FlagSkipUpgradeHeight) {
			j := i + 1

			for j < len(args) {
				tArg := args[j]
				if strings.HasPrefix(tArg, "-") {
					break
				}
				h, err := strconv.Atoi(tArg)
				if err == nil {
					heights = append(heights, h)
				}
				j++
			}

			break
		}
	}
	return heights
}
