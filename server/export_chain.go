package server

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
	bc "github.com/tendermint/tendermint/blockchain"
	"github.com/tendermint/tendermint/node"

	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	flagExportChainHeightStart = "height-start"
	flagExportChainHeightEnd   = "height-end"
)

type exportTx struct {
	Height     int64    `json:"height"`
	Proposer   string   `json:"proposer"`
	Validators []string `json:"validators"`
	Txs        []sdk.Tx `json:"txs"`
}

// ExportChainCmd returns a command that allows for blockchain transaction
// exporting.
func ExportChainCmd(ctx *Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-chain <file>",
		Short: "Export blockchain transactions and metadata by height to a JSON file",
		Long:  ``, // TODO: ...
		Args:  cobra.ExactArgs(1),
		RunE:  exportChainExec(ctx, cdc),
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().Uint64(flagExportChainHeightStart, 1, "Start block height for export")
	cmd.Flags().Uint64(flagExportChainHeightEnd, 0, "End block height for export")

	return cmd
}

func exportChainExec(ctx *Context, cdc *codec.Codec) utils.CobraExecErrFn {
	return func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		startHeight := viper.GetInt64(flagExportChainHeightStart)
		endHeight := viper.GetInt64(flagExportChainHeightEnd)

		// initiate block storage
		blockStoreDB, err := node.DefaultDBProvider(&node.DBContext{
			ID:     "blockstore",
			Config: ctx.Config,
		})
		if err != nil {
			return err
		}

		// create blockstore and get the latest block height
		blockStore := bc.NewBlockStore(blockStoreDB)
		latestHeight := blockStore.Height()

		if err := validateHeights(startHeight, endHeight, latestHeight); err != nil {
			return err
		}

		// create and open file handle for export
		fileHandle, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		defer fileHandle.Close()

		// wrap with a gzip stream if the export file contains the appropriate extension
		var writer io.Writer = fileHandle
		if strings.HasSuffix(filePath, ".gz") {
			writer = gzip.NewWriter(writer)
			defer writer.(*gzip.Writer).Close()
		}

		if endHeight == 0 {
			endHeight = latestHeight
		}

		return exportChain(cdc, startHeight, endHeight, blockStore, writer)
	}
}

func validateHeights(startHeight, endHeight, currentHeight int64) error {
	switch {
	case currentHeight == 0:
		return errors.New("no blocks found")
	case startHeight > currentHeight:
		return fmt.Errorf("invalid block start height: %d", startHeight)
	case endHeight != 0 && startHeight > endHeight:
		return fmt.Errorf("invalid block end height: %d", endHeight)
	}

	return nil
}

func exportChain(cdc *codec.Codec, sHeight, eHeight int64, bs *bc.BlockStore, w io.Writer) error {
	currHeight := sHeight
	txDecoder := auth.DefaultTxDecoder(cdc)
	streamEncoder := json.NewEncoder(w)

	for ; currHeight <= eHeight; currHeight++ {
		block := bs.LoadBlock(currHeight)
		if block == nil {
			return fmt.Errorf("unexpected nil block at height %d", currHeight)
		}

		exportTx := exportTx{
			Height:     block.Height,
			Proposer:   fmt.Sprintf("%X", block.Header.ProposerAddress),
			Validators: make([]string, len(block.LastCommit.Precommits)),
			Txs:        make([]sdk.Tx, len(block.Txs)),
		}

		for i, valAddr := range block.LastCommit.Precommits {
			exportTx.Validators[i] = fmt.Sprintf("%X", valAddr)
		}

		for i, tx := range block.Txs {
			stdTx, err := txDecoder(tx)
			if err != nil {
				return err
			}

			exportTx.Txs[i] = stdTx
		}

		streamEncoder.Encode(exportTx)
	}

	return nil
}
