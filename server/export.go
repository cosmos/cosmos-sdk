package server

// DONTCOVER

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	flagHeight        = "height"
	flagForZeroHeight = "for-zero-height"
	flagJailWhitelist = "jail-whitelist"
)

// ExportCmd dumps app state to JSON.
func ExportCmd(ctx *Context, cdc codec.JSONMarshaler, appExporter AppExporter) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(flags.FlagHome))

			traceWriterFile := viper.GetString(flagTraceStore)

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

			traceWriter, err := openTraceWriter(traceWriterFile)
			if err != nil {
				return err
			}

			height := viper.GetInt64(flagHeight)
			forZeroHeight := viper.GetBool(flagForZeroHeight)
			jailWhiteList := viper.GetStringSlice(flagJailWhitelist)

			appState, validators, cp, err := appExporter(ctx.Logger, db, traceWriter, height, forZeroHeight, jailWhiteList)
			if err != nil {
				return fmt.Errorf("error exporting state: %v", err)
			}

			doc, err := tmtypes.GenesisDocFromFile(ctx.Config.GenesisFile())
			if err != nil {
				return err
			}

			doc.AppState = appState
			doc.Validators = validators
			doc.ConsensusParams = &tmtypes.ConsensusParams{
				Block: tmtypes.BlockParams{
					MaxBytes: cp.Block.MaxBytes,
					MaxGas:   cp.Block.MaxGas,
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

			fmt.Println(string(sdk.MustSortJSON(encoded)))
			return nil
		},
	}

	cmd.Flags().Int64(flagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().Bool(flagForZeroHeight, false, "Export state to start at height zero (perform preproccessing)")
	cmd.Flags().StringSlice(flagJailWhitelist, []string{}, "List of validators to not jail state export")

	return cmd
}
