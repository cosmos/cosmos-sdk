# Trace Plugin

This plugin demonstrates how to listen to state changes of individual `KVStores` as described in [ADR-038 State Listening](https://github.com/vulcanize/cosmos-sdk/blob/adr038_plugin_proposal/docs/architecture/adr-038-state-listening.md).



<!-- TOC -->
- [Running the plugin](#running-the-plugin)
- [Plugin design](#plugin-design)


## Running the plugin

The plugin is setup to run as the `default` plugin. See `./plugin/loader/preload_list` for how to enable and disable default plugins. For lighter unit test run: `./plugin/plugins/kafka/service/service_test.go`. 

1. Copy the content below to `~/app.toml`.

   ```
   # app.toml

   . . .

   ###############################################################################
   ###                      Plugin system configuration                        ###
   ###############################################################################

   [plugins]

   # turn the plugin system, as a whole, on or off
   on = true

   # List of plugin names to enable from the plugin/plugins/*
   enabled = ["kafka"]

   # The directory to load non-preloaded plugins from; defaults $GOPATH/src/github.com/cosmos/cosmos-sdk/plugin/plugins
   dir = ""

   ###############################################################################
   ###                       Trace Plugin configuration                        ###
   ###############################################################################

   # The specific parameters for the trace streaming service plugin
   [plugins.streaming.trace]

   # List of store keys we want to expose for this streaming service.
   keys = []

   # In addition to block event info, print the data to stdout as well.
   print_data_to_stdout = false

   # Whether or not to halt the application when plugin fails to deliver message(s).
   halt_app_on_delivery_error = false
   ```

2. Run `make test-sim-nondeterminism-state-listening-trace` and wait for the tests to finish.


## Plugin design
The plugin is an example implementation of [ADR-038 State Listening](https://docs.cosmos.network/master/architecture/adr-038-state-listening.html) where state change events get logged at `DEBUG` level.
