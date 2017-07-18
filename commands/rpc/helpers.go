package rpc

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin/commands"

	"github.com/tendermint/tendermint/rpc/client"
)

var waitCmd = &cobra.Command{
	Use:   "wait",
	Short: "Wait until a given height, or number of new blocks",
	RunE:  commands.RequireInit(runWait),
}

func init() {
	waitCmd.Flags().Int(FlagHeight, -1, "wait for block height")
	waitCmd.Flags().Int(FlagDelta, -1, "wait for given number of nodes")
}

func runWait(cmd *cobra.Command, args []string) error {
	c := commands.GetNode()
	h := viper.GetInt(FlagHeight)
	if h == -1 {
		// read from delta
		d := viper.GetInt(FlagDelta)
		if d == -1 {
			return errors.New("Must set --height or --delta")
		}
		status, err := c.Status()
		if err != nil {
			return err
		}
		h = status.LatestBlockHeight + d
	}

	// now wait
	err := client.WaitForHeight(c, h, nil)
	if err != nil {
		return err
	}
	fmt.Printf("Chain now at height %d\n", h)
	return nil
}
