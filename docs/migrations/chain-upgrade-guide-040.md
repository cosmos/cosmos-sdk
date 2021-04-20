<!--
order: 2
-->

# Chain Upgrade Guide to v0.42

This document explains how to perform a chain upgrade from v0.39 to v0.42. {synopsis}

::: tip
Please note that the three SDK versions v0.40, v0.41 and v0.42 are functionally equivalent, together called the "Stargate" series. The version bumps are consequences of post-release state-breaking bugfixes.
:::

## Risks

As a validator, performing the upgrade procedure on your consensus nodes carries a heightened risk of double-signing and
being slashed: if your validator node votes for a block, and, in the same block time, restarts the upgraded node, this may lead to double-voting on a block.

The riskiest thing a validator can do is to discover that they made a mistake and repeat the upgrade procedure again during
the network startup. If you discover a mistake in the process, the best thing to do is wait for the network to start
before correcting it. If the network is halted and you have started with a different genesis file than the expected one,
seek advice from the validator community.

## Recovery

- Prior to exporting the state, the validators are encouraged to take a full data snapshot at exported height. Exported
  height will be determined by a governance proposal. Data backup is usually done by copying daemon home directory,
  e.g.: `~/.simd`

**Note:** we use "simd" as our app throughout this doc, be sure to replace with the name of your own binary.

It is critically important to back-up the validator state file, e.g.: `~/.simd/data/priv_validator_state.json` file
after stopping your daemon process. This file is updated every block as your validator participates in a consensus
rounds. It is a critical file needed to prevent double-signing, in case the upgrade fails, and the previous chain needs
to be restarted.

In the event that the upgrade does not succeed, validators and operators must downgrade back to old version of the
software and restore to their latest snapshot before restarting their nodes.

## Upgrade procedure

1. The procedure is to export the state from the old binary, and import it with the new binary. First, verify your old binary version (which should use `github.com/cosmos/cosmos-sdk@0.39.*`) before exporting the state.

   ```shell
   simd version --long
   ```

1. Export the state from existing chain using the old binary.

   ```shell
   simd export --for-zero-height --height <height> > v039_exported_state.json
   ```

1. Verify the SHA256 of the (sorted) exported genesis file:

   ```shell
   $ jq -S -c -M '' v039_exported_state.json | shasum -a 256
   [SHASUM_PLACEHOLDER]  v039_exported_state.json
   ```

1. Cross check the hash with other peers (other validators) in the chat rooms.

1. Install the latest binary (which uses `github.com/cosmos/cosmos-sdk@0.40.*`).

1. Migrate the exported state to `github.com/cosmos/cosmos-sdk@0.40.*` compatible genesis state.

   ```shell
   simd migrate v0.42 v039_exported_state.json --chain-id <new_chain_id> --genesis-time <new_genesis_time_in_utc> > new_v042_genesis.json
   ```

   **Note:** The migrate command takes an input genesis state and migrates it to a targeted version. New `genesis-time` is usually mentioned in the governance proposal, and should be passed as flag argument. If the flag is omitted, then the genesis time of the upgraded chain will be the same as the old one, which may cause confusion.

1. All the necessary state changes are handled in the `simd migrate v0.42` migration command. However, Tendermint parameters are **not** handled in this command. You might need to update these parameters manually.

   In the recent versions of Tendermint, the following changes have been made:

   - `consensus_params.evidence.max_num` has been renamed to `consensus_params.evidence.max_bytes`.
   - `consensus_params.evidence.max_age` has been removed, and replaced by `consensus_params.evidence.max_age_duration` and `consensus_params.evidence.max_age_num_blocks`.

   Make sure that your genesis JSON files contains the correct values specific to your chain. If the `simd migrate` errors with a message saying that the genesis file cannot be parsed, these are the fields to check first.

1. Verify the SHA256 of the migrated genesis file with other validators to make sure there are no manual errors in the process.

   ```shell
   $ jq -S -c -M '' new_v042_genesis.json | shasum -a 256
   [SHASUM_PLACEHOLDER]  new_v042_genesis.json
   ```

1. Make sure to update the genesis parameters in the new genesis if any. All these details will be generally present in
   the governance proposal.

1) If your chain is using IBC, make sure to add IBC initial genesis state to the genesis file. You can use the following command to add IBC initial genesis state to the genesis file.

   ```shell
   cat new_v042_genesis.json | jq '.app_state |= . + {"ibc":{"client_genesis":{"clients":[],"clients_consensus":[],"create_localhost":false},"connection_genesis":{"connections":[],"client_connection_paths":[]},"channel_genesis":{"channels":[],"acknowledgements":[],"commitments":[],"receipts":[],"send_sequences":[],"recv_sequences":[],"ack_sequences":[]}},"transfer":{"port_id":"transfer","denom_traces":[],"params":{"send_enabled":false,"receive_enabled":false}},"capability":{"index":"1","owners":[]}}' > tmp_genesis.json && mv tmp_genesis.json new_v042_genesis.json
   ```

   **Note:** This would add IBC state with IBC's `send_enabled: false` and `receive_enabled: false`. Make sure to update them to `true` in the above command if are planning to enable IBC transactions with chain upgrade. Otherwise you can do it via a governance proposal.

1) Reset the old state.

   **Note:** Be sure you have a complete backed up state of your node before proceeding with this step.
   See Recovery for details on how to proceed.

   ```shell
   simd unsafe-reset-all
   ```

1) Move the new genesis.json to your daemon config directory. Ex

   ```shell
   cp new_v042_genesis.json ~/.simd/config/genesis.json
   ```

1) Update `~/.simd/config/app.toml` to include latest app configurations. [Here is the link](https://github.com/cosmos/cosmos-sdk/blob/v0.42.0-rc6/server/config/toml.go#L11-L164) to the default template for v0.42's `app.toml`. Make sure to
   update your custom configurations as per your validator design, e.g. `gas_price`.

   Compared to v0.39, some notable updates to `app.toml` are:

   - API server is now configured to run in-process with daemon, previously it was a separate process, invoked by running rest-server
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

1) Kill if any external `rest-server` process is running.

1) All set now! You can (re)start your daemon to validate on the upgraded network. Make sure to check your binary version
   before starting the daemon:

   ```
   simd version --long
   ```

## Next {hide}

Once your chain is upgraded, make sure to [update your clients' REST endpoints](./rest.md).
