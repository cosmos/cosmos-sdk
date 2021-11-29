package plugin

import (
	"io"

	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
)

const (
	// PLUGINS_SYMBOL is the symbol for loading Cosmos-SDK plugins from a linked .so file
	PLUGINS_SYMBOL = "Plugins"

	// PLUGINS_TOML_KEY is the top-level TOML key for plugin configuration
	PLUGINS_TOML_KEY = "plugins"

	// PLUGINS_ON_TOML_KEY is the second-level TOML key for turning on the plugin system as a whole
	PLUGINS_ON_TOML_KEY = "on"

	// PLUGINS_DIR_TOML_KEY is the second-level TOML key for the directory to load plugins from
	PLUGINS_DIR_TOML_KEY = "dir"

	// PLUGINS_DISABLED_TOML_KEY is the second-level TOML key for a list of plugins to disable
	PLUGINS_DISABLED_TOML_KEY = "disabled"

	// DEFAULT_PLUGINS_DIRECTORY is the default directory to load plugins from
	DEFAULT_PLUGINS_DIRECTORY = "src/github.com/cosmos/cosmos-sdk/plugin/plugins"
)

// Plugin is the base interface for all kinds of cosmos-sdk plugins
// It will be included in interfaces of different Plugins
type Plugin interface {
	// Name should return unique name of the plugin
	Name() string

	// Version returns current version of the plugin
	Version() string

	// Init is called once when the Plugin is being loaded
	// The plugin is passed the AppOptions for configuration
	// A plugin will not necessarily have a functional Init
	Init(env serverTypes.AppOptions) error

	// Closer interface for shutting down the plugin process
	io.Closer
}
