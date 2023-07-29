package rpc

import (
	"context"
	"encoding/json"

	"github.com/cometbft/cometbft/p2p"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// ValidatorInfo is info about the node's validator, same as CometBFT,
// except that we use our own PubKey.
type validatorInfo struct {
	Address     []byte             `json:"address"`
	PubKey      cryptotypes.PubKey `json:"pub_key"`
	VotingPower int64              `json:"voting_power"`
}

// ResultStatus is node's info, same as CometBFT, except that we use our own
// PubKey.
type resultStatus struct {
	NodeInfo      p2p.DefaultNodeInfo `json:"node_info"`
	SyncInfo      coretypes.SyncInfo  `json:"sync_info"`
	ValidatorInfo validatorInfo       `json:"validator_info"`
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

			status, err := getNodeStatus(clientCtx)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey
			// `status` has TM pubkeys, we need to convert them to our pubkeys.
			if status.ValidatorInfo.PubKey != nil {
				pk, err = cryptocodec.FromCmtPubKeyInterface(status.ValidatorInfo.PubKey)
				if err != nil {
					return err
				}
			}
			statusWithPk := resultStatus{
				NodeInfo: status.NodeInfo,
				SyncInfo: status.SyncInfo,
				ValidatorInfo: validatorInfo{
					Address:     status.ValidatorInfo.Address,
					PubKey:      pk,
					VotingPower: status.ValidatorInfo.VotingPower,
				},
			}

			output, err := json.Marshal(statusWithPk)
			if err != nil {
				return err
			}

			return clientCtx.PrintRaw(output)
		},
	}

	cmd.Flags().StringP(flags.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	cmd.Flags().StringP(flags.FlagOutput, "o", "json", "Output format (text|json)")

	return cmd
}

func getNodeStatus(clientCtx client.Context) (*coretypes.ResultStatus, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return &coretypes.ResultStatus{}, err
	}

	return node.Status(context.Background())
}
