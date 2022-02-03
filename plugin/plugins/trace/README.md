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

    # This is a TOML config file.
    # For more information, see https://github.com/toml-lang/toml
    
    ###############################################################################
    ###                           Base Configuration                            ###
    ###############################################################################
    
    # Impose a global wait limit threshold for ListenSuccess() messages of external streaming services. (seconds)
    # It is recomended to set this higher then the average block commit time.
    globalWaitLimit = 30
    
    
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
    # external streaming service before signaling back to `app.Commit()` call.
    # This threshold is used to synchronize the work between `app.Commit()` and the
    # `ABCIListener.ListenSuccess()` call. `ListenSucess()` will allow up to the
    # specified threshold for services to complete writing messages. The completion
    # is signaled when `ListenEndBlock` has finished writting.
    # This value MUST BE less than the 'globalWaitLimit' threshold as not to trigger
    # the 'globalWaitLimit' timeout which will halt the app.
    deliveredBlockTimeoutSeconds = 2
    ```
   
2. Run `make test-sim-nondeterminism` and wait for the tests to finish.


## Plugin design
The plugin is an example implementation of [ADR-038 State Listening](https://docs.cosmos.network/master/architecture/adr-038-state-listening.html) where state change events get logged at `DEBUG` level.
