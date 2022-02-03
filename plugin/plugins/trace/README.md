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

   ...
    
   ###############################################################################
   ###                      Plugin system configuration                        ###
   ###############################################################################
    
   [plugins]
    
   # turn the plugin system, as a whole, on or off
   on = true
    
   # list of plugins to disable
   disabled = []
    
   # The directory to load non-preloaded plugins from; defaults to
   dir = ""
    
   # a mapping of plugin-specific streaming service parameters, mapped to their pluginFileName
   [plugins.streaming]
    
   ###############################################################################
   ###                       Trace Plugin configuration                        ###
   ###############################################################################
   
   # The specific parameters for the Kafka streaming service plugin
   [plugins.streaming.trace]
   
   # List of store keys we want to expose for this streaming service.
   keys = []
   
   # Timeout threshold for which a particular block's messages must be delivered to
   # external streaming service before signaling back to the `ack` channel.
   # If the `ack` is set to `false` this setting will be ignored.
   # Note: This setting MUST be less then `plugins.global_ack_wait_limit`.
   #       Otherwise, the application will halt without committing blocks.
   # In milliseconds.
   deliver_block_wait_limit = 2000
   
   # In addition to block event info, print the data to stdout as well.
   print_data_to_stdout = false
   
   # whether to operate in fire-and-forget or success/failure acknowledgement mode
   # false == fire-and-forget; true == sends a message receipt success/fail signal
   ack = "false"
    ```
   
2. Run `make test-sim-nondeterminism` and wait for the tests to finish.


## Plugin design
The plugin is an example implementation of [ADR-038 State Listening](https://docs.cosmos.network/master/architecture/adr-038-state-listening.html) where state change events get logged at `DEBUG` level.
