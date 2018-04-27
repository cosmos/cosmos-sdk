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
func ExportCmd(app baseapp.AppExporter, ctx *Context) *cobra.Command {
	export := exportCmd{
		appExporter: app,
		context:     ctx,
	}
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE:  export.run,
	}
	return cmd
}

type exportCmd struct {
	appExporter baseapp.AppExporter
	context     *Context
}

func (e exportCmd) run(cmd *cobra.Command, args []string) error {
	home := viper.GetString("home")
	genesis, cdc, err := e.appExporter(home, e.context.Logger)
	if err != nil {
		return errors.Errorf("Error exporting state: %v\n", err)
	}
	output, err := wire.MarshalJSONIndent(cdc, genesis)
	if err != nil {
		return errors.Errorf("Error marshalling state: %v\n", err)
	}
	fmt.Println(string(output))
	return nil
}
