package rpc

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/tendermint/tendermint/libs/bytes"
	"github.com/tendermint/tendermint/p2p"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// ValidatorInfo is info about the node's validator, same as Tendermint,
// except that we use our own PubKey.
type validatorInfo struct {
	Address     bytes.HexBytes
	PubKey      cryptotypes.PubKey
	VotingPower int64
}

// ResultStatus is node's info, same as Tendermint, except that we use our own
// PubKey.
type resultStatus struct {
	NodeInfo      nodeInfo
	SyncInfo      coretypes.SyncInfo
	ValidatorInfo validatorInfo
}

type nodeInfo struct {
	ProtocolVersion p2p.ProtocolVersion      `json:"protocol_version"`
	DefaultNodeID   p2p.ID                   `json:"default_node_id,omitempty"`
	ListenAddr      string                   `json:"listen_addr,omitempty"`
	Network         string                   `json:"network,omitempty"`
	Version         string                   `json:"version,omitempty"`
	Channels        bytes.HexBytes           `json:"channels,omitempty"`
	Moniker         string                   `json:"moniker,omitempty"`
	Other           p2p.DefaultNodeInfoOther `json:"other"`
	BinaryVersion   string                   `json:"binary_version"`
}

// StatusCommand returns the command to return the status of the network.
func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query remote node for status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := tmservice.NewServiceClient(clientCtx)
			res, err := queryClient.GetNodeInfo(context.Background(), &tmservice.GetNodeInfoRequest{})
			if err != nil {
				return err
			}

			status, err := getNodeStatus(clientCtx)
			if err != nil {
				return err
			}

			// `status` has TM pubkeys, we need to convert them to our pubkeys.
			pk, err := cryptocodec.FromTmPubKeyInterface(status.ValidatorInfo.PubKey)
			if err != nil {
				return err
			}

			infoWithBinary := nodeInfo{
				ProtocolVersion: status.NodeInfo.ProtocolVersion,
				DefaultNodeID:   status.NodeInfo.DefaultNodeID,
				ListenAddr:      status.NodeInfo.ListenAddr,
				Network:         status.NodeInfo.Network,
				Version:         status.NodeInfo.Version,
				Channels:        status.NodeInfo.Channels,
				Moniker:         status.NodeInfo.Moniker,
				Other:           status.NodeInfo.Other,
				BinaryVersion:   res.ApplicationVersion.GetVersion(),
			}

			statusWithPk := resultStatus{
				NodeInfo: infoWithBinary,
				SyncInfo: status.SyncInfo,
				ValidatorInfo: validatorInfo{
					Address:     status.ValidatorInfo.Address,
					PubKey:      pk,
					VotingPower: status.ValidatorInfo.VotingPower,
				},
			}

			output, err := clientCtx.LegacyAmino.MarshalJSON(statusWithPk)
			if err != nil {
				return err
			}

			cmd.Println(string(output))
			return nil
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")

	return cmd
}

func getNodeStatus(clientCtx client.Context) (*coretypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &coretypes.ResultStatus{}, err
	}

	return node.Status(context.Background())
}
