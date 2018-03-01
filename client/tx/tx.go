package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client" // XXX: not good
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

// Get the default command for a tx query
func QueryTxCmd(cmdr commander) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "Matches this txhash over all committed blocks",
		RunE:  cmdr.queryTxCmd,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	return cmd
}

// command to query for a transaction
func (c commander) queryTxCmd(cmd *cobra.Command, args []string) error {
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
	node, err := client.GetNode()
	if err != nil {
		return err
	}
	prove := !viper.GetBool(client.FlagTrustNode)

	res, err := node.Tx(hash, prove)
	if err != nil {
		return err
	}
	info, err := formatTxResult(c.cdc, res)
	if err != nil {
		return err
	}

	output, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

func formatTxResult(cdc *wire.Codec, res *ctypes.ResultTx) (txInfo, error) {
	// TODO: verify the proof if requested
	tx, err := parseTx(cdc, res.Tx)
	if err != nil {
		return txInfo{}, err
	}

	info := txInfo{
		Height: res.Height,
		Tx:     tx,
		Result: res.TxResult,
	}
	return info, nil
}

// txInfo is used to prepare info to display
type txInfo struct {
	Height int64                  `json:"height"`
	Tx     sdk.Tx                 `json:"tx"`
	Result abci.ResponseDeliverTx `json:"result"`
}

func parseTx(cdc *wire.Codec, txBytes []byte) (sdk.Tx, error) {
	var tx sdk.StdTx
	err := cdc.UnmarshalBinary(txBytes, &tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
