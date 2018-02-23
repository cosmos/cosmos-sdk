package tx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

const (
	flagTags = "tag"
	flagAny  = "any"
)

func txSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Search for all transactions that match the given tags",
		RunE:  searchTx,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	cmd.Flags().StringSlice(flagTags, nil, "Tags that must match (may provide multiple)")
	cmd.Flags().Bool(flagAny, false, "Return transactions that match ANY tag, rather than ALL")
	return cmd
}

func searchTx(cmd *cobra.Command, args []string) error {
	tags := viper.GetStringSlice(flagTags)
	if len(tags) == 0 {
		return errors.New("Must declare at least one tag to search")
	}
	// XXX: implement ANY
	query := strings.Join(tags, " AND ")

	// get the node
	uri := viper.GetString(client.FlagNode)
	if uri == "" {
		return errors.New("Must define which node to query with --node")
	}
	node := client.GetNode(uri)

	prove := !viper.GetBool(client.FlagTrustNode)
	res, err := node.TxSearch(query, prove)
	if err != nil {
		return err
	}

	info, err := formatTxResults(res)
	if err != nil {
		return err
	}

	cdc := app.MakeTxCodec()
	output, err := cdc.MarshalJSON(info)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

func formatTxResults(res []*ctypes.ResultTx) ([]txInfo, error) {
	var err error
	out := make([]txInfo, len(res))
	for i := range res {
		out[i], err = formatTxResult(res[i])
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
