package rpc

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
)

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  checkStatus,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func checkStatus(cmd *cobra.Command, args []string) error {
	// get the node
	uri := viper.GetString(client.FlagNode)
	node := client.GetNode(uri)
	res, err := node.Status()
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(res, "  ", "")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
