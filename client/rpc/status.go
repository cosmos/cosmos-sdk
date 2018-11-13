package rpc

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/spf13/viper"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// StatusCommand returns the status of the network
func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE:  printNodeStatus,
	}

	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	viper.BindPFlag(client.FlagNode, cmd.Flags().Lookup(client.FlagNode))
	cmd.Flags().Bool(client.FlagIndentResponse, false, "Add indent to JSON response")
	return cmd
}

func getNodeStatus(cliCtx context.CLIContext) (*ctypes.ResultStatus, error) {
	// get the node
	node, err := cliCtx.GetNode()
	if err != nil {
		return &ctypes.ResultStatus{}, err
	}

	return node.Status()
}

// CMD

func printNodeStatus(cmd *cobra.Command, args []string) error {
	// No need to verify proof in getting node status
	viper.Set(client.FlagTrustNode, true)
	cliCtx := context.NewCLIContext()
	status, err := getNodeStatus(cliCtx)
	if err != nil {
		return err
	}

	var output []byte
	if cliCtx.Indent {
		output, err = cdc.MarshalJSONIndent(status, "", "  ")
	} else {
		output, err = cdc.MarshalJSON(status)
	}
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

// REST

// REST handler for node info
func NodeInfoRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(cliCtx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		nodeInfo := status.NodeInfo
		utils.PostProcessResponse(w, cdc, nodeInfo, cliCtx.Indent)
	}
}

// REST handler for node syncing
func NodeSyncingRequestHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(cliCtx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		syncing := status.SyncInfo.CatchingUp
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(strconv.FormatBool(syncing)))
	}
}
