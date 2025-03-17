# cosmossdk.io/tools/benchmark

A benchmark module to test chain and storage performance. It can be used to holistically test
the end to end performance of a node. Given an initial configuration tools/benchmark provides:

* A possibly enormous sequence of key-value sets in InitGenesis distributed across n storekeys,
  e.g. 20M keys across 5 store keys
* A client which syncs to genesis state then deterministically generates txs which contain a
  configurable sequence of get, insert, update & delete operations
* A keeper which processes the above transactions and emits some telemetry data about them.

Client invocation looks like:

```bash
simdv2 tx benchmark load-test --from bob --yes --ops 1000 --pause 10 -v
```

On exit it dumps the generator state so that running again should still be in sync. It assumes
that any transaction accepted by the network was processed, which may not be the case, so miss
rate will probably increase over time. This isn't really a problem for tests.

Obviously this module is built to DOS a node by testing the upper bounds of chain performance;
when testing gas limits should be increased. It should not be included in chains by default but
is enabled in simapp for testing.