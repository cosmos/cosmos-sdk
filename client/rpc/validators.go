package rpc

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	wire "github.com/tendermint/go-wire"

	"github.com/cosmos/cosmos-sdk/client"
)

func validatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validatorset <height>",
		Short: "Get the full validator set at given height",
		RunE:  getValidators,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	return cmd
}

func getValidators(cmd *cobra.Command, args []string) error {
	var height *int64
	// optional height
	if len(args) > 0 {
		h, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}
		if h > 0 {
			tmp := int64(h)
			height = &tmp
		}
	}

	// get the node
	node, err := client.GetNode()
	if err != nil {
		return err
	}

	res, err := node.Validators(height)
	if err != nil {
		return err
	}

	output, err := wire.MarshalJSON(res)
	// output, err := json.MarshalIndent(res, "  ", "")
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
