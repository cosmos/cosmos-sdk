package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/server"
)

func main() {
	cmd := &cobra.Command{
		Use:   "mock_node",
		Short: "A mock node for testing cosmovisor.",
		Long: `The --halt-interval flag is required and must be specified in order to halt the node.
The --upgrade-plan and --halt-height flags are mutually exclusive. It is an error to specify both.
Based on which flag is specified the node will either exhibit --halt-height before or
x/upgrade upgrade-info.json behavior.`,
	}
	var haltInterval time.Duration
	var upgradePlan string
	var haltHeight uint64
	cmd.Flags().DurationVar(&haltInterval, "halt-interval", 0, "Interval to wait before halting the node")
	cmd.Flags().StringVar(&upgradePlan, "upgrade-plan", "", "upgrade-info.json to create after the halt duration is reached. Either this flag or --halt-height must be specified but not both.")
	cmd.Flags().Uint64Var(&haltHeight, server.FlagHaltHeight, 0, "Block height at which to gracefully halt the chain and shutdown the node. E")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if upgradePlan != "" && haltHeight > 0 {
			return fmt.Errorf("cannot specify both --upgrade-plan and --halt-height")
		}
		if upgradePlan == "" && haltHeight == 0 {
			return fmt.Errorf("must specify either --upgrade-plan or --halt-height")
		}
		if haltInterval == 0 {
			return fmt.Errorf("must specify --halt-interval")
		}
		time.Sleep(haltInterval)
		if haltHeight > 0 {
			panic(fmt.Errorf("halt per configuration height %d time %d", haltHeight, 0))
		} else if upgradePlan != "" {
			panic("upgrade-info.json not implemented")
		}
		return nil
	}
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
