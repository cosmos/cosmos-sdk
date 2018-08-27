package server

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/wire"
	tmtypes "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"path"
)

// ExportCmd dumps app state to JSON.
func ExportCmd(ctx *Context, cdc *wire.Codec, appExporter AppExporter) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := viper.GetString("home")
			traceStore := viper.GetString(flagTraceStore)
			emptyState, err := isEmptyState(home)
			if err != nil {
				return err
			}

			if emptyState {
				fmt.Println("WARNING: State is not initialized. Returning genesis file.")
				genesisFile := path.Join(home, "config", "genesis.json")
				genesis, err := ioutil.ReadFile(genesisFile)
				if err != nil {
					return err
				}
				fmt.Println(string(genesis))
				return nil
			}

			appState, validators, err := appExporter(home, ctx.Logger, traceStore)
			if err != nil {
				return errors.Errorf("error exporting state: %v\n", err)
			}

			doc, err := tmtypes.GenesisDocFromFile(ctx.Config.GenesisFile())
			if err != nil {
				return err
			}

			doc.AppState = appState
			doc.Validators = validators

			encoded, err := wire.MarshalJSONIndent(cdc, doc)
			if err != nil {
				return err
			}

			fmt.Println(string(encoded))
			return nil
		},
	}
}

func isEmptyState(home string) (bool, error) {
	files, err := ioutil.ReadDir(path.Join(home, "data"))
	if err != nil {
		return false, err
	}

	return len(files) == 0, nil
}
