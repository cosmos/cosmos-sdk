package server

import (
	"compress/gzip"
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
)

const (
	flagExportChainHeightStart = "height-start"
	flagExportChainHeightEnd   = "height-end"
)

// TODO: ...
func ExportChainCmd(ctx *Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export-chain <file>",
		Short: "Export blockchain transactions and metadata by height to a JSON file",
		Long:  ``, // TODO: ...
		Args:  cobra.ExactArgs(1),
		RunE:  exportChainExec(ctx),
	}

	cmd.Flags().Uint64(flagExportChainHeightStart, 0, "Start block height for export")
	cmd.Flags().Uint64(flagExportChainHeightEnd, 0, "End block height for export")

	return cmd
}

func exportChainExec(ctx *Context) utils.CobraExecErrFn {
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

		return exportChain(startHeight, endHeight, blockStore, writer)
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

func exportChain(startHeight, endHeight int64, blockStore *bc.BlockStore, w io.Writer) error {
	var currHeight int64
	currHeight = startHeight

	// type exportChain struct {
	// 	Height uint64 `json:"height"`
	// }

	// json.NewEncoder(w)

	for ; currHeight < endHeight; currHeight++ {
		block := blockStore.LoadBlock(currHeight)
		for _, tx := range block.Txs {

		}
	}

	return nil
}
