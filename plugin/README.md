# Comsos-SDK Plugins
This package contains an extensible plugin system for the Cosmos-SDK. Included in this top-level package is the base interface
for a Cosmos-SDK plugin, as well as more specific plugin interface definitions that build on top of this base interface.
The [loader](./loader) sub-directory contains the Go package and scripts for loading plugins into the SDK. The [plugins](./plugins)
sub-directory contains the preloaded plugins and a script for building them, this is also the directory that the plugin loader will look
for non-preloaded plugins by default.

The base plugin interface is defined as:
```go
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

	// Closer interface to shutting down the plugin process
	io.Closer
}
```

Specific plugin types extend this interface, enabling them to work with the loader tooling defined in the [loader sub-directory](./loader).

The plugin system itself is configured using the `plugins` TOML mapping in the App's `app.toml` file. There are three
parameters for configuring the plugins: `plugins.on`, `plugins.enabled` and `plugins.dir`. `plugins.on` is a bool that
turns on or off the plugin system at large, `plugins.dir` directs the system to a directory to load plugins from, and
`plugins.enabled` is a list enabled plugin names.

```toml
[plugins]
    on = false # turn the plugin system, as a whole, on or off
    enabled = ["list", "of", "plugin", "names", "to", "enable"]
    dir = "the directory to load non-preloaded plugins from; defaults to cosmos-sdk/plugin/plugins"
```

As mentioned above, some plugins can be preloaded. This means they do not need to be loaded from the specified `plugins.dir` and instead
are loaded by default. Note, both preloaded and non-preloaded plugins must appear in `plugins.enabled` list for the app to send events to them.
This provides node operators with the ability to `opt-in` and enable only plugins of interest. At this time the only preloaded plugins are;
the [file streaming service plugin](./plugins/file), the [trace streaming service plugin](./plugins/trace) and the [kafka streaming service plugin](./plugins/kafka).
Plugins can be added to the preloaded set by adding the plugin to the [plugins dir](../../plugin/plugin.go) and modifying the [preload_list](../../plugin/loader/preload_list).

In your application, if the  `plugins.on` is set to `true` use this to direct the invocation of `NewPluginLoader` and walk through
the steps of plugin loading, initialization, injection, starting, and closure.

e.g. in `NewSimApp`:

```go
func NewSimApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig simappparams.EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *SimApp {

	...

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, ibchost.StoreKey, upgradetypes.StoreKey,
		evidencetypes.StoreKey, ibctransfertypes.StoreKey, capabilitytypes.StoreKey,
	)

	pluginsOnKey := fmt.Sprintf("%s.%s", plugin.PLUGINS_TOML_KEY, plugin.PLUGINS_ON_TOML_KEY)
	if cast.ToBool(appOpts.Get(pluginsOnKey)) {
		// this loads the preloaded and any plugins found in `plugins.dir`
		// if their names match those in the `plugins.enabled` list.
		pluginLoader, err := loader.NewPluginLoader(appOpts, logger)
		if err != nil {
			// handle error
		}

		// initialize the loaded plugins
		if err := pluginLoader.Initialize(); err != nil {
			// handle error
		}

		// register the plugin(s) with the BaseApp
		if err := pluginLoader.Inject(bApp, appCodec, keys); err != nil {
			// handle error
		}

		// start the plugin services, optionally use wg to synchronize shutdown using io.Closer
		wg := new(sync.WaitGroup)
		if err := pluginLoader.Start(wg); err != nil {
			// handler error
		}
	}

	...

	return app
}
```

# State Streaming Plugin
The `BaseApp` package contains the interface for a `StreamingService` used to write state changes out from individual KVStores to a
file or stream, as described in [ADR-038](../docs/architecture/adr-038-state-listening.md).

Specific `StreamingService` implementations are written and loaded as plugins by extending the above interface with a
`StateStreamingPlugin` interface that adds `Register` method used to register the plugin's `StreamingService` with the
`BaseApp` and a `Start` method to start the streaming service.

```go
// StateStreamingPlugin interface for plugins that load a streaming.Service onto a baseapp.BaseApp
type StateStreamingPlugin interface {
	// Register configures and registers the plugin streaming service with the BaseApp
	Register(bApp *baseapp.BaseApp, marshaller codec.BinaryCodec, keys map[string]*types.KVStoreKey) error

	// Start starts the background streaming process of the plugin streaming service
	Start(wg *sync.WaitGroup) error

	// Plugin is the base Plugin interface
	Plugin
}
```

A `StateStreamingPlugin` is configured from within an App using the `AppOptions` loaded from the `app.toml` file.
Every `StateStreamingPlugin` will be configured within the `plugins.streaming` TOML mapping. The exact keys/parameters
present in this mapping will be dependent on the specific `StateStreamingPlugin`, but we will introduce some standards
here using the file `StateStreamingPlugin`:

Plugin TOML configuration should be split into separate sub-tables for each kind of plugin (e.g. `plugins.streaming`).

Within these sub-tables, the parameters for a specific plugin of that kind are included in another sub-table (e.g. `plugins.streaming.file`).
It is generally expected, but not required, that a streaming service plugin can be configured with a set of store keys
(e.g. `plugins.streaming.file.keys`) for the stores it listens to and a flag (e.g. `plugins.streaming.file.halt_app_on_delivery_error`)
that signifies whether the service operates in a fire-and-forget capacity, or the BaseApp should halt in case of a delivery error by the plugin service.
The file `StreamingService` does not have an individual `halt_app_on_delivery_error` since it operates synchronously with the App.

e.g.

```toml
[plugins]
    on = false # turn the plugin system, as a whole, on or off
    enabled = ["list", "of", "plugin", "names", "to", "enable"]
    dir = "the directory to load non-preloaded plugins from; defaults to cosmos-sdk/plugin/plugins"
    [plugins.streaming] # a mapping of plugin-specific streaming service parameters, mapped to their pluginFileName
        [plugins.streaming.file] # the specific parameters for the file streaming service plugin
            keys = ["list", "of", "store", "keys", "we", "want", "to", "expose", "for", "this", "streaming", "service"]
            write_dir = "path to the write directory"
            prefix = "optional prefix to prepend to the generated file names
            # Whether or not to halt the application when plugin fails to deliver message(s).
            halt_app_on_delivery_error = false # false = fire-and-forget
```
