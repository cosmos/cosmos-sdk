package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/tendermint/p2p"
)

// ShowNodeIDCmd - ported from Tendermint, dump node ID to stdout
func ShowNodeIDCmd(ctx *Context) *cobra.Command {
	cmd := showNodeId{ctx}
	return &cobra.Command{
		Use:   "show_node_id",
		Short: "Show this node's ID",
		RunE:  cmd.run,
	}
}

type showNodeId struct {
	context *Context
}

func (s showNodeId) run(cmd *cobra.Command, args []string) error {
	cfg := s.context.Config
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}
