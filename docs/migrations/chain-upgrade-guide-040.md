# Upgrading a live chain from 039 to 040

### Risks
As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and
being slashed. The most important piece of this procedure is verifying your software version and genesis file hash before
starting your validator and signing.

The riskiest thing a validator can do is discover that they made a mistake and repeat the upgrade procedure again during
the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start
before correcting it. If the network is halted and you have started with a different genesis file than the expected one,
seek advice from the validator community.

### Recovery
* Prior to exporting the state, the validators are encouraged to take a full data snapshot at exported height. Exported
height will be determined by a governance proposal. Data backup is usually done by copying daemon home directory,
ex: `~/.simd`
  
**Note:** we use "simd" as our app throughout this doc, be sure to replace with the name of your own binary

It is critically important to back-up the validator state file, ex: `~/.simd/data/priv_validator_state.json` file
after stopping your daemon process. This file is updated every block as your validator participates in a consensus
rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails, and the previous chain needs
to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to old version of the
software and restore to their latest snapshot before restarting their nodes.

### Upgrade procedure
- Use old binary to export the state. Make sure to verify your binary version before exporting the state
- Export the state from existing chain using old-binary (which uses `sdk@0.39.x`).
Example:
    ```sh
    simd export --for-zero-height --height <height> > 039_exported_state.json
    ```
- Verify the SHA256 of the (sorted) exported genesis file:
    ```shell
    $ jq -S -c -M '' 039_exported_state.json | shasum -a 256
    [PLACEHOLDER]  039_exported_state.json
    ```
- Cross check the hash with other peers (other validators) in the chat rooms
- Install the latest binary (which uses `0.40`)
- Migrate the exported state to `0.40` compatible genesis state
    ```shell
    simd migrate v40 039_exported_state.json --chain-id <new_chain_id> --genesis-time <new_genesis_time_in_utc> > new_v40_genesis.json
    ```
  NOTE: The migrate command takes an input genesis state and migrates it to a targeted version. New `genesis-time` will
  be as mentioned in the governance proposal.
- Verify the SHA256 of the migrated genesis file with other valdiators to make sure there are no manual errors in the process.
    ```shell
    $ jq -S -c -M '' new_v40_genesis.json | shasum -a 256
    [PLACEHOLDER]  new_v40_genesis.json
    ```
- Make sure to update the genesis parameters in the new genesis if any. All these details will be generally present in
the governance proposal
- All the necessary state chanegs are handled in `040` migration command, including tendermint params. So, manual updates to
the genesis are not required
- If your chain is using IBC, make sure to add IBC initiaal genesis state to the genesis file. You can use the following
  command to add IBC initial genesis state to the genesis file.
  ```shell
  cat new_040_genesis.json | jq '.app_state |= . + {"ibc":{"client_genesis":{"clients":[],"clients_consensus":[],"create_localhost":false},"connection_genesis":{"connections":[],"client_connection_paths":[]},"channel_genesis":{"channels":[],"acknowledgements":[],"commitments":[],"receipts":[],"send_sequences":[],"recv_sequences":[],"ack_sequences":[]}},"transfer":{"port_id":"transfer","denom_traces":[],"params":{"send_enabled":false,"receive_enabled":false}},"capability":{"index":"1","owners":[]}}' > new_040_genesis.json
  ```
  Note: This would add ibc state with ibc's `send_enabled: false` and `receive_enabled: false`. Make sure to update them to `true` in the above command if are planning to enable IBC transactions with chain upgrade. Otherwise you can do it via a governance proposal.
- Reset the old state
NOTE: Be sure you have a complete backed up state of your node before proceeding with this step.
See Recovery for details on how to proceed.
```shell
simd unsafe-reset-all
```
- Move the new genesis.json to your daemon config directory. Ex
```shell script
cp new_v40_genesis.json ~/.simd/config/genesis.json
```

- Update `~/.simd/config/app.toml` to include latest app configurations. Below is an example `app.toml`. Make sure to
update your custom configurations as per your validator design ex: `gas_price`.

```yaml
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml
###############################################################################
###                           Base Configuration                            ###
###############################################################################
# The minimum gas prices a validator is willing to accept for processing a
# transaction. A transaction's fees must meet the minimum of any denomination
# specified in this config (e.g. 0.25token1;0.0001token2).
minimum-gas-prices = ""
# default: the last 100 states are kept in addition to every 500th state; pruning at 10 block intervals
# nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
# everything: all saved states will be deleted, storing only the current state; pruning at 10 block intervals
# custom: allow pruning options to be manually specified through 'pruning-keep-recent', 'pruning-keep-every', and 'pruning-interval'
pruning = "default"
# These are applied if and only if the pruning strategy is custom.
pruning-keep-recent = "0"
pruning-keep-every = "0"
pruning-interval = "0"
# HaltHeight contains a non-zero block height at which a node will gracefully
# halt and shutdown that can be used to assist upgrades and testing.
#
# Note: Commitment of state will be attempted on the corresponding block.
halt-height = 0
# HaltTime contains a non-zero minimum block time (in Unix seconds) at which
# a node will gracefully halt and shutdown that can be used to assist upgrades
# and testing.
#
# Note: Commitment of state will be attempted on the corresponding block.
halt-time = 0
# MinRetainBlocks defines the minimum block height offset from the current
# block being committed, such that all blocks past this offset are pruned
# from Tendermint. It is used as part of the process of determining the
# ResponseCommit.RetainHeight value during ABCI Commit. A value of 0 indicates
# that no blocks should be pruned.
#
# This configuration value is only responsible for pruning Tendermint blocks.
# It has no bearing on application state pruning which is determined by the
# "pruning-*" configurations.
#
# Note: Tendermint block pruning is dependant on this parameter in conunction
# with the unbonding (safety threshold) period, state pruning and state sync
# snapshot parameters to determine the correct minimum value of
# ResponseCommit.RetainHeight.
min-retain-blocks = 0
# InterBlockCache enables inter-block caching.
inter-block-cache = true
# IndexEvents defines the set of events in the form {eventType}.{attributeKey},
# which informs Tendermint what to index. If empty, all events will be indexed.
#
# Example:
# ["message.sender", "message.recipient"]
index-events = []
###############################################################################
###                         Telemetry Configuration                         ###
###############################################################################
[telemetry]
# Prefixed with keys to separate services.
service-name = ""
# Enabled enables the application telemetry functionality. When enabled,
# an in-memory sink is also enabled by default. Operators may also enabled
# other sinks such as Prometheus.
enabled = false
# Enable prefixing gauge values with hostname.
enable-hostname = false
# Enable adding hostname to labels.
enable-hostname-label = false
# Enable adding service to labels.
enable-service-label = false
# PrometheusRetentionTime, when positive, enables a Prometheus metrics sink.
prometheus-retention-time = 0
# GlobalLabels defines a global set of name/value label tuples applied to all
# metrics emitted using the wrapper functions defined in telemetry package.
#
# Example:
# [["chain_id", "cosmoshub-1"]]
global-labels = [
]
###############################################################################
###                           API Configuration                             ###
###############################################################################
[api]
# Enable defines if the API server should be enabled.
enable = false
# Swagger defines if swagger documentation should automatically be registered.
swagger = false
# Address defines the API server to listen on.
address = "tcp://0.0.0.0:1317"
# MaxOpenConnections defines the number of maximum open connections.
max-open-connections = 1000
# RPCReadTimeout defines the Tendermint RPC read timeout (in seconds).
rpc-read-timeout = 10
# RPCWriteTimeout defines the Tendermint RPC write timeout (in seconds).
rpc-write-timeout = 0
# RPCMaxBodyBytes defines the Tendermint maximum response body (in bytes).
rpc-max-body-bytes = 1000000
# EnableUnsafeCORS defines if CORS should be enabled (unsafe - use it at your own risk).
enabled-unsafe-cors = false
###############################################################################
###                           gRPC Configuration                            ###
###############################################################################
[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true
# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
###############################################################################
###                        State Sync Configuration                         ###
###############################################################################
# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]
# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 0
# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```

The updates to `app.toml` are:
- API is now configured to run in-process with daemon, previously it was a separate process, invoked by running rest-server
command i.e., `gaiacli rest-server`. Now it is in-process with daemon and can be enabled/disabled by API configuration:
```yaml
[api]
# Enable defines if the API server should be enabled.
enable = false
# Swagger defines if swagger documentation should automatically be registered.
swagger = false
```
`swagger` setting refers to enabling/disabling swagger docs API, i.e, `/swagger/` API endpoint.

- gRPC Configuration
```yaml
[grpc]
# Enable defines if the gRPC server should be enabled.
enable = true
# Address defines the gRPC server address to bind to.
address = "0.0.0.0:9090"
```

- State Sync Configuration
```yaml
# State sync snapshots allow other nodes to rapidly join the network without replaying historical
# blocks, instead downloading and applying a snapshot of the application state at a given height.
[state-sync]
# snapshot-interval specifies the block interval at which local state sync snapshots are
# taken (0 to disable). Must be a multiple of pruning-keep-every.
snapshot-interval = 0
# snapshot-keep-recent specifies the number of recent snapshots to keep and serve (0 to keep all).
snapshot-keep-recent = 2
```

- Kill if any external `rest-server` process is running

- All set now! You can (re)start your daemon to validate on the upgraded network. Make sure to check your binary version
before starting the daemon. Ex:

```
simd version --long
```
