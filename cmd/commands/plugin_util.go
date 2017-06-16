package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tendermint/basecoin/types"
)

type plugin struct {
	name      string
	newPlugin func() types.Plugin
}

var plugins = []plugin{}

// RegisterStartPlugin is used to enable a plugin
func RegisterStartPlugin(name string, newPlugin func() types.Plugin) {
	plugins = append(plugins, plugin{name: name, newPlugin: newPlugin})
}

//Returns a version command based on version input
func QuickVersionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}
