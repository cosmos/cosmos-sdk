package rpc

import (
	"github.com/spf13/cobra"
	"github.com/tendermint/basecoin/client/commands"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Query the status of the node",
	RunE:  commands.RequireInit(runStatus),
}

func runStatus(cmd *cobra.Command, args []string) error {
	c := commands.GetNode()
	status, err := c.Status()
	if err != nil {
		return err
	}
	return printResult(status)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Query info on the abci app",
	RunE:  commands.RequireInit(runInfo),
}

func runInfo(cmd *cobra.Command, args []string) error {
	c := commands.GetNode()
	info, err := c.ABCIInfo()
	if err != nil {
		return err
	}
	return printResult(info)
}

var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Query the genesis of the node",
	RunE:  commands.RequireInit(runGenesis),
}

func runGenesis(cmd *cobra.Command, args []string) error {
	c := commands.GetNode()
	genesis, err := c.Genesis()
	if err != nil {
		return err
	}
	return printResult(genesis)
}

var validatorsCmd = &cobra.Command{
	Use:   "validators",
	Short: "Query the validators of the node",
	RunE:  commands.RequireInit(runValidators),
}

func runValidators(cmd *cobra.Command, args []string) error {
	c := commands.GetNode()
	validators, err := c.Validators()
	if err != nil {
		return err
	}
	return printResult(validators)
}
