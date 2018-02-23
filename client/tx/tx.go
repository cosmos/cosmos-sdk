package tx

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app" // XXX: not good
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx <hash>",
		Short: "Matches this txhash over all committed blocks",
		RunE:  queryTx,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	return cmd
}

func queryTx(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a tx hash")
	}

	// find the key to look up the account
	hexStr := args[0]
	hash, err := hex.DecodeString(hexStr)
	if err != nil {
		return err
	}

	// get the node
	uri := viper.GetString(client.FlagNode)
	if uri == "" {
		return errors.New("Must define which node to query with --node")
	}
	node := client.GetNode(uri)
	prove := !viper.GetBool(client.FlagTrustNode)

	res, err := node.Tx(hash, prove)
	if err != nil {
		return err
	}
	info, err := formatTxResult(res)
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

func formatTxResult(res *ctypes.ResultTx) (txInfo, error) {
	height := res.Height
	result := res.TxResult
	// TODO: verify the proof if requested

	tx, err := parseTx(res.Tx)
	if err != nil {
		return txInfo{}, err
	}

	info := txInfo{
		Height: height,
		Tx:     tx,
		Result: result,
	}
	return info, nil
}

// txInfo is used to prepare info to display
type txInfo struct {
	Height int64                  `json:"height"`
	Tx     sdk.Tx                 `json:"tx"`
	Result abci.ResponseDeliverTx `json:"result"`
}

func parseTx(txBytes []byte) (sdk.Tx, error) {
	var tx sdk.StdTx
	cdc := app.MakeTxCodec()
	err := cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
