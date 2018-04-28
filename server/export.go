package server

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/wire"
)

// ExportCmd dumps app state to JSON
func ExportCmd(ctx *Context, cdc *wire.Codec, appExporter baseapp.AppExporter) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := viper.GetString("home")
			appState, err := appExporter(home, ctx.Logger)
			if err != nil {
				return errors.Errorf("Error exporting state: %v\n", err)
			}
			fmt.Println(string(output))
			return nil
		},
	}
}
