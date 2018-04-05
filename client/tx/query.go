package tx

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	abci "github.com/tendermint/abci/types"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Get the default command for a tx query
func QueryTxCmd(cmdr commander) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx [hash]",
		Short: "Matches this txhash over all committed blocks",
		RunE:  cmdr.queryAndPrintTx,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:46657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	return cmd
}

func (c commander) queryTx(hashHexStr string, trustNode bool) ([]byte, error) {
	hash, err := hex.DecodeString(hashHexStr)
	if err != nil {
		return nil, err
	}

	// get the node
	node, err := context.NewCoreContextFromViper().GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.Tx(hash, !trustNode)
	if err != nil {
		return nil, err
	}
	info, err := formatTxResult(c.cdc, res)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(info, "", "  ")
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

// CMD

// command to query for a transaction
func (c commander) queryAndPrintTx(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide a tx hash")
	}

	// find the key to look up the account
	hashHexStr := args[0]
	trustNode := viper.GetBool(client.FlagTrustNode)

	output, err := c.queryTx(hashHexStr, trustNode)
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}

// REST

func QueryTxRequestHandler(cdc *wire.Codec) func(http.ResponseWriter, *http.Request) {
	c := commander{cdc}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		hashHexStr := vars["hash"]
		trustNode, err := strconv.ParseBool(r.FormValue("trust_node"))
		// trustNode defaults to true
		if err != nil {
			trustNode = true
		}

		output, err := c.queryTx(hashHexStr, trustNode)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}
