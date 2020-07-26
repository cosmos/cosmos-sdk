package rpc

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	"github.com/KiraCore/cosmos-sdk/codec/legacy"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	"github.com/KiraCore/cosmos-sdk/version"

	"github.com/tendermint/tendermint/p2p"
)

// StatusCommand returns the command to return the status of the network.
func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			status, err := getNodeStatus(clientCtx)
			if err != nil {
				return err
			}

			output, err := legacy.Cdc.MarshalJSON(status)
			if err != nil {
				return err
			}

			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")

	return cmd
}

func getNodeStatus(clientCtx client.Context) (*ctypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &ctypes.ResultStatus{}, err
	}

	return node.Status()
}

// NodeInfoResponse defines a response type that contains node status and version
// information.
type NodeInfoResponse struct {
	p2p.DefaultNodeInfo `json:"node_info"`

	ApplicationVersion version.Info `json:"application_version"`
}

// REST handler for node info
func NodeInfoRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(clientCtx)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		resp := NodeInfoResponse{
			DefaultNodeInfo:    status.NodeInfo,
			ApplicationVersion: version.NewInfo(),
		}

		rest.PostProcessResponseBare(w, clientCtx, resp)
	}
}

// SyncingResponse defines a response type that contains node syncing information.
type SyncingResponse struct {
	Syncing bool `json:"syncing"`
}

// REST handler for node syncing
func NodeSyncingRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, err := getNodeStatus(clientCtx)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		rest.PostProcessResponseBare(w, clientCtx, SyncingResponse{Syncing: status.SyncInfo.CatchingUp})
	}
}
