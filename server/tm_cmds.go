package server

import (
	"fmt"

	"github.com/spf13/cobra"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/p2p"
	pvm "github.com/tendermint/tendermint/types/priv_validator"
)

// ShowNodeIDCmd - ported from Tendermint, dump node ID to stdout
func ShowNodeIDCmd(ctx *Context) *cobra.Command {
	cmd := showNodeID{ctx}
	return &cobra.Command{
		Use:   "show_node_id",
		Short: "Show this node's ID",
		RunE:  cmd.run,
	}
}

type showNodeID struct {
	context *Context
}

func (s showNodeID) run(cmd *cobra.Command, args []string) error {
	cfg := s.context.Config
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}

//--------------------------------------------------------------------------------

// ShowValidator - ported from Tendermint, show this node's validator info
func ShowValidatorCmd(ctx *Context) *cobra.Command {
	cmd := showValidator{ctx}
	return &cobra.Command{
		Use:   "show_validator",
		Short: "Show this node's validator info",
		RunE:  cmd.run,
	}
}

type showValidator struct {
	context *Context
}

func (s showValidator) run(cmd *cobra.Command, args []string) error {
	cfg := s.context.Config
	privValidator := pvm.LoadOrGenFilePV(cfg.PrivValidatorFile())
	pubKeyJSONBytes, err := cdc.MarshalJSON(privValidator.PubKey)
	if err != nil {
		return err
	}
	fmt.Println(string(pubKeyJSONBytes))
	return nil
}

//------------------------------------------------------------------------------

// UnsafeResetAllCmd - extension of the tendermint command, resets initialization
func UnsafeResetAllCmd(ctx *Context) *cobra.Command {
	cmd := resetAll{ctx}
	return &cobra.Command{
		Use:   "unsafe_reset_all",
		Short: "Reset all blockchain data",
		RunE:  cmd.run,
	}
}

type resetAll struct {
	context *Context
}

func (r resetAll) run(cmd *cobra.Command, args []string) error {
	cfg := r.context.Config
	tcmd.ResetAll(cfg.DBDir(), cfg.PrivValidatorFile(), r.context.Logger)
	return nil
}
