---
sidebar_position: 1
---

# `SimApp`

`SimApp` is an application built using the Cosmos SDK for testing and educational purposes.

## Running testnets with `simd`

If you want to spin up a quick testnet with your friends, you can follow these steps.
Unless otherwise noted, every step must be done by everyone who wants to participate
in this testnet.

1. From the root directory of the Cosmos SDK repository, run `$ make build`. This will build the
    `simd` binary inside a new `build` directory. The following instructions are run from inside
    the `build` directory.
2. If you've run `simd` before, you may need to reset your database before starting a new
    testnet. You can reset your database with the following command: `$ ./simd comet unsafe-reset-all`.
3. `$ ./simd init [moniker] --chain-id [chain-id]`. This will initialize a new working directory
    at the default location `~/.simapp`. You need to provide a "moniker" and a "chain id". These
    two names can be anything, but you will need to use the same "chain id" in the following steps.
4. `$ ./simd keys add [key_name]`. This will create a new key, with a name of your choosing.
    Save the output of this command somewhere; you'll need the address generated here later.
5. `$ ./simd genesis add-genesis-account [key_name] [amount]`, where `key_name` is the same key name as
    before; and `amount` is something like `10000000000000000000000000stake`.
6. `$ ./simd genesis gentx [key_name] [amount] --chain-id [chain-id]`. This will create the genesis
    transaction for your new chain. Here `amount` should be at least `1000000000stake`. If you
    provide too much or too little, you will encounter an error when starting your node.
7. Now, one person needs to create the genesis file `genesis.json` using the genesis transactions
   from every participant, by gathering all the genesis transactions under `config/gentx` and then
   calling `$ ./simd genesis collect-gentxs`. This will create a new `genesis.json` file that includes data
   from all the validators (we sometimes call it the "super genesis file" to distinguish it from
   single-validator genesis files).
8. Once you've received the super genesis file, overwrite your original `genesis.json` file with
    the new super `genesis.json`.
9. Modify your `config/config.toml` (in the simapp working directory) to include the other participants as
    persistent peers:

    ```text
    # Comma separated list of nodes to keep persistent connections to
    persistent_peers = "[validator_address]@[ip_address]:[port],[validator_address]@[ip_address]:[port]"
    ```

    You can find `validator_address` by running `$ ./simd comet show-node-id`. The output will
    be the hex-encoded `validator_address`. The default `port` is 26656.
10. Now you can start your nodes: `$ ./simd start`.

Now you have a small testnet that you can use to try out changes to the Cosmos SDK or CometBFT!

NOTE: Sometimes creating the network through the `collect-gentxs` will fail, and validators will start
in a funny state (and then panic). If this happens, you can try to create and start the network first
with a single validator and then add additional validators using a `create-validator` transaction.

## Determinism simulation guide

The `simapp` simulator includes an extended determinism test:

* `TestAppStateDeterminismExtended` in `simapp/sim_test.go`

This test replays randomized simulations and compares full per-block app-hash traces.
For replay runs (run 2..N), app hashes are checked incrementally at each finalized block
against run 1 and fail fast on the first mismatch.

Simulation tx execution now always runs through the default BaseApp lifecycle:

* `CheckTx` (mempool admission)
* `PrepareProposal`
* `ProcessProposal`
* finalize/deliver execution

This keeps simulator execution closer to real node behavior than direct deliver-only paths.
The simulator summary also includes tx lifecycle failure breakdowns by phase
(`checktx`, `prepare`, `process`, `finalize`) with per-reason counts.
It also prints:

* factory pre-delivery skips total (message factory pre-delivery skips)
* lifecycle rejects total (txs rejected in `checktx`/`prepare`/`process`/`finalize`)

### Summary controls and exports

* `SIMAPP_SUMMARY_TOP_N`: when set to `N > 0`, limits lifecycle and skip-reason output to top N entries per section and buckets the remainder as `other`.
* `SIMAPP_SUMMARY_EXPORT_DIR`: when set, each run exports:
    * `seed-<seed>.summary.json`
    * `seed-<seed>.summary.csv`
  containing execution counts/skip reasons and lifecycle phase breakdowns.

### What does "runs per seed" mean?

`SIMAPP_EXTENDED_RUNS_PER_SEED` controls how many times the simulator runs the exact same seed.

Example:

* `SIMAPP_EXTENDED_RUNS_PER_SEED=3`
* `-Seed=932727706452219696`

means:

* one seed is selected (from `-Seed` or randomly)
* that seed is replayed 3 times
* traces/hashes from runs 2..N are compared against run 1 for the same seed

If a seed would lead to an empty validator set, the harness retries candidate seeds automatically.

### Recommended command (heavy)

Run from the repository root:

```bash
cd simapp && \
SIMAPP_EXTENDED_DETERMINISM=1 \
SIMAPP_EXTENDED_NUM_BLOCKS=5000 \
SIMAPP_EXTENDED_BLOCK_SIZE=1200 \
SIMAPP_EXTENDED_RUNS_PER_SEED=5 \
SIMAPP_EXTENDED_PROGRESS_EVERY=5 \
SIMAPP_EXTENDED_SEED_RETRIES=5 \
SIMAPP_SUMMARY_TOP_N=20 \
SIMAPP_SUMMARY_EXPORT_DIR=./sim-summaries \
go test -tags sims . -run TestAppStateDeterminismExtended -count=1 -v -timeout 0 \
  -Seed=932727706452219696
```

### Environment variables

* `SIMAPP_EXTENDED_DETERMINISM`: set to `1` to enable this test
* `SIMAPP_EXTENDED_NUM_BLOCKS`: number of blocks per simulation
* `SIMAPP_EXTENDED_BLOCK_SIZE`: operations per block
* `SIMAPP_EXTENDED_RUNS_PER_SEED`: number of replays per seed
* `SIMAPP_EXTENDED_PROGRESS_EVERY`: print progress every N finalized blocks
* `SIMAPP_EXTENDED_SEED_RETRIES`: max candidate-seed retries to avoid empty-validator-set runs
