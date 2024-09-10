package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"cosmossdk.io/tools/cosmovisor"
)

func NewBatchAddUpgradeCmd() *cobra.Command {
	addUpgrade := &cobra.Command{
		Use:          "add-batch-upgrade <upgrade1-name>:<path-to-exec1> <upgrade2-name>:<path-to-exec2> .. <upgradeN-name>:<path-to-execN>",
		Short:        "Add APP upgrades binary to cosmovisor",
		SilenceUsage: true,
		Args:         cobra.ArbitraryArgs,
		RunE:         AddBatchUpgrade,
	}

	addUpgrade.Flags().Bool(cosmovisor.FlagForce, false, "overwrite existing upgrade binary / upgrade-info.json file")
	addUpgrade.Flags().Int64(cosmovisor.FlagUpgradeHeight, 0, "define a height at which to upgrade the binary automatically (without governance proposal)")

	return addUpgrade
}

func AddBatchUpgrade(cmd *cobra.Command, args []string) error {
	for i, as := range args {
		a := strings.Split(as, ":")
		if len(a) != 2 {
			return fmt.Errorf("argument at position %d (%s) is invalid", i, as)
		}
		if err := AddUpgrade(cmd, a); err != nil {
			return err
		}
	}
	return nil
}
