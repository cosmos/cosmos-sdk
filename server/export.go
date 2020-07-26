package server

// DONTCOVER

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	"github.com/KiraCore/cosmos-sdk/codec"
	sdk "github.com/KiraCore/cosmos-sdk/types"
)

const (
	flagHeight        = "height"
	flagForZeroHeight = "for-zero-height"
	flagJailWhitelist = "jail-whitelist"
)

// ExportCmd dumps app state to JSON.
func ExportCmd(appExporter AppExporter, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.JSONMarshaler

			serverCtx := GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			config.SetRoot(homeDir)

			db, err := openDB(config.RootDir)
			if err != nil {
				return err
			}

			if appExporter == nil {
				if _, err := fmt.Fprintln(os.Stderr, "WARNING: App exporter not defined. Returning genesis file."); err != nil {
					return err
				}

				genesis, err := ioutil.ReadFile(config.GenesisFile())
				if err != nil {
					return err
				}

				fmt.Println(string(genesis))
				return nil
			}

			traceWriterFile, _ := cmd.Flags().GetString(flagTraceStore)
			traceWriter, err := openTraceWriter(traceWriterFile)
			if err != nil {
				return err
			}

			height, _ := cmd.Flags().GetInt64(flagHeight)
			forZeroHeight, _ := cmd.Flags().GetBool(flagForZeroHeight)
			jailWhiteList, _ := cmd.Flags().GetStringSlice(flagJailWhitelist)

			appState, validators, cp, err := appExporter(serverCtx.Logger, db, traceWriter, height, forZeroHeight, jailWhiteList)
			if err != nil {
				return fmt.Errorf("error exporting state: %v", err)
			}

			doc, err := tmtypes.GenesisDocFromFile(serverCtx.Config.GenesisFile())
			if err != nil {
				return err
			}

			doc.AppState = appState
			doc.Validators = validators
			doc.ConsensusParams = &tmtypes.ConsensusParams{
				Block: tmtypes.BlockParams{
					MaxBytes:   cp.Block.MaxBytes,
					MaxGas:     cp.Block.MaxGas,
					TimeIotaMs: doc.ConsensusParams.Block.TimeIotaMs,
				},
				Evidence: tmtypes.EvidenceParams{
					MaxAgeNumBlocks: cp.Evidence.MaxAgeNumBlocks,
					MaxAgeDuration:  cp.Evidence.MaxAgeDuration,
				},
				Validator: tmtypes.ValidatorParams{
					PubKeyTypes: cp.Validator.PubKeyTypes,
				},
			}

			encoded, err := codec.MarshalJSONIndent(cdc, doc)
			if err != nil {
				return err
			}

			cmd.Println(string(sdk.MustSortJSON(encoded)))
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().Int64(flagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().Bool(flagForZeroHeight, false, "Export state to start at height zero (perform preproccessing)")
	cmd.Flags().StringSlice(flagJailWhitelist, []string{}, "List of validators to not jail state export")

	return cmd
}
