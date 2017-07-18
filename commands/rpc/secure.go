package rpc

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/basecoin/commands"
)

func init() {
	blockCmd.Flags().Int(FlagHeight, -1, "block height")
	commitCmd.Flags().Int(FlagHeight, -1, "block height")
	headersCmd.Flags().Int(FlagMin, -1, "minimum block height")
	headersCmd.Flags().Int(FlagMax, -1, "maximum block height")
}

var blockCmd = &cobra.Command{
	Use:   "block",
	Short: "Get a validated block at a given height",
	RunE:  commands.RequireInit(runBlock),
}

func runBlock(cmd *cobra.Command, args []string) error {
	c, err := getSecureNode()
	if err != nil {
		return err
	}

	h := viper.GetInt(FlagHeight)
	block, err := c.Block(h)
	if err != nil {
		return err
	}
	return printResult(block)
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Get the header and commit signature at a given height",
	RunE:  commands.RequireInit(runCommit),
}

func runCommit(cmd *cobra.Command, args []string) error {
	c, err := getSecureNode()
	if err != nil {
		return err
	}

	h := viper.GetInt(FlagHeight)
	commit, err := c.Commit(h)
	if err != nil {
		return err
	}
	return printResult(commit)
}

var headersCmd = &cobra.Command{
	Use:   "headers",
	Short: "Get all headers in the given height range",
	RunE:  commands.RequireInit(runHeaders),
}

func runHeaders(cmd *cobra.Command, args []string) error {
	c, err := getSecureNode()
	if err != nil {
		return err
	}

	min := viper.GetInt(FlagMin)
	max := viper.GetInt(FlagMax)
	headers, err := c.BlockchainInfo(min, max)
	if err != nil {
		return err
	}
	return printResult(headers)
}
