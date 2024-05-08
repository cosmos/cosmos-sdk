package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"cosmossdk.io/errors"

	corectx "cosmossdk.io/core/context"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

const flagGenTxDir = "gentx-dir"

// CollectGenTxsCmd - return the cobra command to collect genesis transactions
func CollectGenTxsCmd(genBalIterator types.GenesisBalancesIterator, validator types.MessageValidator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect-gentxs",
		Short: "Collect genesis txs and output a genesis.json file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			serverCtx := corectx.GetServerContextFromCmd(cmd)
			config, ok := serverCtx.GetConfig().(server.CometConfig)
			if !ok {
				return fmt.Errorf("Can not convert cometbft config")
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			nodeID, valPubKey, err := genutil.InitializeNodeValidatorFiles(config.Config)
			if err != nil {
				return errors.Wrap(err, "failed to initialize node validator files")
			}

			appGenesis, err := types.AppGenesisFromFile(config.GenesisFile())
			if err != nil {
				return errors.Wrap(err, "failed to read genesis doc from file")
			}

			genTxDir, _ := cmd.Flags().GetString(flagGenTxDir)
			genTxsDir := genTxDir
			if genTxsDir == "" {
				genTxsDir = filepath.Join(config.RootDir, "config", "gentx")
			}

			toPrint := newPrintInfo(config.Moniker, appGenesis.ChainID, nodeID, genTxsDir, json.RawMessage(""))
			initCfg := types.NewInitConfig(appGenesis.ChainID, genTxsDir, nodeID, valPubKey)

			appMessage, err := genutil.GenAppStateFromConfig(cdc, clientCtx.TxConfig, config.Config, initCfg, appGenesis, genBalIterator, validator, clientCtx.ValidatorAddressCodec, clientCtx.AddressCodec)
			if err != nil {
				return errors.Wrap(err, "failed to get genesis app state from config")
			}

			toPrint.AppMessage = appMessage

			return displayInfo(toPrint)
		},
	}

	cmd.Flags().String(flagGenTxDir, "", "override default \"gentx\" directory from which collect and execute genesis transactions; default [--home]/config/gentx/")

	return cmd
}
