package server

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

// ExportCmd dumps app state to JSON
func ExportCmd(ctx *Context, cdc *wire.Codec, appExporter AppExporter) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := viper.GetString("home")
			appState, validators, err := appExporter(home, ctx.Logger)
			if err != nil {
				return errors.Errorf("Error exporting state: %v\n", err)
			}
			doc, err := tmtypes.GenesisDocFromFile(ctx.Config.GenesisFile())
			if err != nil {
				return err
			}
			doc.AppStateJSON = appState
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
