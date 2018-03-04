package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  printNodeStatus,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	return cmd
}

func getNodeStatus() ([]byte, error) {
	// get the node
	node, err := client.GetNode()
	if err != nil {
		return nil, err
	}
	res, err := node.Status()
	if err != nil {
		return nil, err
	}

	output, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return nil, err
	}

	return output, nil
}

// CMD

func printNodeStatus(cmd *cobra.Command, args []string) error {
	status, err := getNodeStatus()
	if err != nil {
		return err
	}
	fmt.Println(string(status))
	return nil
}

// REST

// TODO match desired spec output
func NodeStatusRequestHandler(w http.ResponseWriter, r *http.Request) {
	status, err := getNodeStatus()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
	w.Write(status)
}
