package server

import (
	"fmt"

	"github.com/spf13/cobra"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tmlibs/log"
)

// ShowNodeIdCmd - ported from Tendermint, dump node ID to stdout
func ShowNodeIdCmd(logger log.Logger) *cobra.Command {
	cmd := showNodeId{logger}
	return &cobra.Command{
		Use:   "show_node_id",
		Short: "Show this node's ID",
		RunE:  cmd.run,
	}
}

type showNodeId struct {
	logger log.Logger
}

func (s showNodeId) run(cmd *cobra.Command, args []string) error {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	if err != nil {
		return err
	}
	fmt.Println(nodeKey.ID())
	return nil
}
