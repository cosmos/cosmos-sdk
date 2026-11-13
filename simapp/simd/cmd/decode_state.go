package cmd

import (
	"fmt"
	"path/filepath"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/spf13/cobra"

	"cosmossdk.io/simapp"

	"github.com/cosmos/cosmos-sdk/server"
)

// decodeStateCmd decodes every key and value of the collections-backed module
// stores at the latest committed height of a stopped node, and fails on any
// entry that does not decode with the codecs the keepers read it with.
func decodeStateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "decode-state",
		Short: "Decode all collections state of a stopped node and report entries that do not decode",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := server.GetServerContextFromCmd(cmd)
			dataDir := filepath.Join(serverCtx.Config.RootDir, "data")
			db, err := dbm.NewDB("application", server.GetAppDBBackend(serverCtx.Viper), dataDir)
			if err != nil {
				return err
			}
			defer db.Close()

			app := simapp.NewSimApp(serverCtx.Logger, db, true, serverCtx.Viper)
			height := app.LastBlockHeight()
			ctx := app.NewContextLegacy(true, cmtproto.Header{Height: height})
			if err := app.CheckStateDecodes(ctx); err != nil {
				return fmt.Errorf("state at height %d does not decode:\n%w", height, err)
			}
			cmd.Printf("all collections state decodes at height %d\n", height)
			return nil
		},
	}
}
