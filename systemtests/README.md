# System Tests

This package contains the testing framework for black-box system tests. It includes a test runner that sets up a 
multi-node blockchain locally for use in tests. The framework provides utilities and helpers for easy access and 
setup in tests.

## Components

- **CLI**: Command-line interface wrapper for interacting with the chain or keyring
- **Servers**: Server instances to run the blockchain environment.
- **Events**: Event listeners
- **RPC**: Remote Procedure Call setup for communication.

## Dependencies

- **testify**: Testing toolkit.
- **gjson**: JSON parser.
- **sjson**: JSON modifier.

Server and client-side operations are executed on the host machine.

## Developer

### Test strategy

System tests cover the full stack via cli and a running (multi node) network. They are more expensive (in terms of time/ cpu) 
to run compared to unit or integration tests. 
Therefore, we focus on the **critical path** and do not cover every condition.

## How to use

Read the [getting_started.md](../systemtests/getting_started.md) guide to get started.

### Execute a single test

```sh
go test -tags system_test -count=1 -v . --run TestStakeUnstake  -verbose
```

Test cli parameters

* `-verbose` verbose output
* `-wait-time` duration - time to wait for chain events (default 30s)
* `-nodes-count` int - number of nodes in the cluster (default 4)

# Port ranges

With *n* nodes:

* `26657` - `26657+n` - RPC
* `1317` - `1317+n` - API
* `9090` - `9090+n` - GRPC
* `16656` - `16656+n` - P2P

For example Node *3* listens on `26660` for RPC calls

## Resources

* [gjson query syntax](https://github.com/tidwall/gjson#path-syntax)

## Disclaimer

This is based on the system test framework in [wasmd](https://github.com/CosmWasm/wasmd) built by Confio.
