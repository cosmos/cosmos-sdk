package server

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/wire"
)

// AppExporter dumps all app state to JSON-serializable structure
type AppExporter func(home string, log log.Logger) (interface{}, *wire.Codec, error)

// ExportCmd dumps app state to JSON
func ExportCmd(app AppExporter, ctx *Context) *cobra.Command {
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
	appExporter AppExporter
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
