# Running a Node

> TODO: Improve documentation of `gaiad`


## Basics

To start a node:

```shell
$ gaiad start <flags>
```

Options for running the `gaiad` binary are effectively the same as for `tendermint`. 
See `gaiad --help` and the 
[guide to using Tendermint](https://github.com/tendermint/tendermint/blob/master/docs/using-tendermint.md) 
for more details.

## Debugging

Optionally, you can run `gaiad` with `--trace-store` to trace all store operations
to a specified file.

```shell
$ gaiad start <flags> --trace-store=/path/to/trace.out
```

Key/value pairs will be base64 encoded. Additionally, the block number and any
correlated transaction hash will be included as metadata.

e.g.
```json
...
{"operation":"write","key":"ATW6Bu997eeuUeRBwv1EPGvXRfPR","value":"BggEEBYgFg==","metadata":{"blockHeight":12,"txHash":"5AAC197EC45E6C5DE0798C4A4E2F54BBB695CA9E"}}
{"operation":"write","key":"AjW6Bu997eeuUeRBwv1EPGvXRfPRCgAAAAAAAAA=","value":"AQE=","metadata":{"blockHeight":12,"txHash":"5AAC197EC45E6C5DE0798C4A4E2F54BBB695CA9E"}}
{"operation":"read","key":"ATW6Bu997eeuUeRBwv1EPGvXRfPR","value":"BggEEBYgFg==","metadata":{"blockHeight":13}}
{"operation":"read","key":"AjW6Bu997eeuUeRBwv1EPGvXRfPRCwAAAAAAAAA=","value":"","metadata":{"blockHeight":13}}
...
```

You can then query for the various traced operations using a tool like [jq](https://github.com/stedolan/jq).

```shell
$ jq -s '.[] | select((.key=="ATW6Bu997eeuUeRBwv1EPGvXRfPR") and .metadata.blockHeight==14)' /path/to/trace.out
```