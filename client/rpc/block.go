package rpc

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/tendermint/tmlibs/common"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	flagSelect = "select"
)

//BlockCommand returns the verified block data for a given heights
func BlockCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block [height]",
		Short: "Get verified data for a the block at given height",
		Args:  cobra.MaximumNArgs(1),
		RunE:  printBlock,
	}
	cmd.Flags().StringP(client.FlagNode, "n", "tcp://localhost:26657", "Node to connect to")
	// TODO: change this to false when we can
	cmd.Flags().Bool(client.FlagTrustNode, true, "Don't verify proofs for responses")
	cmd.Flags().StringSlice(flagSelect, []string{"header", "tx"}, "Fields to return (header|txs|results)")
	return cmd
}

type OutputTxEntry struct {
	TX   sdk.Tx          `json:"tx"`
	Hash common.HexBytes `json:"hash"`
}

type OutputBlock struct {
	Header     *types.Header      `json:"header"`
	Evidence   types.EvidenceData `json:"evidence"`
	Data       []OutputTxEntry    `json:"data"`
	LastCommit *types.Commit      `json:"last_commit"`
}

type OutputResultBlock struct {
	BlockMeta *types.BlockMeta `json:"block_meta"`
	Block     OutputBlock      `json:"block"`
}

func getBlock(ctx context.CoreContext, height *int64) ([]byte, error) {
	// get the node
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	// TODO: actually honor the --select flag!
	// header -> BlockchainInfo
	// header, tx -> Block
	// results -> BlockResults
	res, err := node.Block(height)
	if err != nil {
		return nil, err
	}

	data := make([]OutputTxEntry, len(res.Block.Data.Txs))
	for i, dataTx := range res.Block.Data.Txs {
		var tx auth.StdTx
		err := cdc.UnmarshalBinary(dataTx, &tx)
		if err != nil {
			return nil, err
		}
		data[i] = OutputTxEntry{
			TX:   tx,
			Hash: dataTx.Hash(),
		}
	}

	outputResultBlock := OutputResultBlock{
		BlockMeta: res.BlockMeta,
		Block: OutputBlock{
			Header:     res.Block.Header,
			Evidence:   res.Block.Evidence,
			Data:       data,
			LastCommit: res.Block.LastCommit,
		},
	}

	// TODO move maarshalling into cmd/rest functions
	// output, err := tmwire.MarshalJSON(res)
	output, err := cdc.MarshalJSON(outputResultBlock)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// get the current blockchain height
func GetChainHeight(ctx context.CoreContext) (int64, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return -1, err
	}
	status, err := node.Status()
	if err != nil {
		return -1, err
	}
	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// CMD

func printBlock(cmd *cobra.Command, args []string) error {
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

	output, err := getBlock(context.NewCoreContextFromViper(), height)
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}

// REST

// REST handler to get a block
func BlockRequestHandlerFn(ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		height, err := strconv.ParseInt(vars["height"], 10, 64)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte("ERROR: Couldn't parse block height. Assumed format is '/block/{height}'."))
			return
		}
		chainHeight, err := GetChainHeight(ctx)
		if height > chainHeight {
			w.WriteHeader(404)
			w.Write([]byte("ERROR: Requested block height is bigger then the chain length."))
			return
		}
		output, err := getBlock(ctx, &height)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}

// REST handler to get the latest block
func LatestBlockRequestHandlerFn(ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		height, err := GetChainHeight(ctx)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		output, err := getBlock(ctx, &height)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(output)
	}
}
