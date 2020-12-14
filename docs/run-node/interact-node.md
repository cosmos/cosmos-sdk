<!--
order: 3
-->

# Interacting with the Node

There are multiple ways to interact with a node: using the CLI, using gRPC or using the REST endpoints. {synopsis}

## Pre-requisite Readings

- [gRPC, REST and Tendermint Endpoints](../core/grpc_rest.md) {prereq}
- [Running a Node](./run-node.md) {prereq}

## Using the CLI

Now that your chain is running, it is time to try sending tokens from the first account you created to a second account. In a new terminal window, start by running the following query command:

```bash
simd query account $MY_VALIDATOR_ADDRESS --chain-id my-test-chain
```

You should see the current balance of the account you created, equal to the original balance of `stake` you granted it minus the amount you delegated via the `gentx`. Now, create a second account:

```bash
simd keys add recipient --keyring-backend test

# Put the generated address in a variable for later use.
RECIPIENT=$(simd keys show recipient -a --keyring-backend test)
```

The command above creates a local key-pair that is not yet registered on the chain. An account is created the first time it receives tokens from another account. Now, run the following command to send tokens to the `recipient` account:

```bash
simd tx send $MY_VALIDATOR_ADDRESS $RECIPIENT 1000stake --chain-id my-test-chain

# Check that the recipient account did receive the tokens.
simd query account $RECIPIENT --chain-id my-test-chain
```

Finally, delegate some of the stake tokens sent to the `recipient` account to the validator:

```bash
simd tx staking delegate $(simd keys show my_validator --bech val -a --keyring-backend test) 500stake --from recipient --chain-id my-test-chain

# Query the total delegations to `validator`.
simd query staking delegations-to $(simd keys show my_validator --bech val -a --keyring-backend test) --chain-id my-test-chain
```

You should see two delegations, the first one made from the `gentx`, and the second one you just performed from the `recipient` account.

## Using gRPC

The Protobuf ecosystem developed tools for different use cases, including code-generation from `*.proto` files into various languages. These tools allow to build clients easily. Often, the client connection (i.e. the transport) can be plugged and replaced very easily. Let's explore one of the most popular transport: [gRPC](../core/grpc_rest.md).

Since the code generation library largely depends on your own tech stack, we will only present two alternatives:

- `grpcurl` for generic debugging and testing,
- CosmJS for JavaScript/TypeScript developers.

### grpcurl: Reflection, Queries, and Simulation

[grpcurl])https://github.com/fullstorydev/grpcurl is like `curl` but for gRPC. It is also available as a Go library, but we will use it only as a CLI command for debugging and testing purposes. Follow the instructions in the previous link to install it.

Assuming you have a local node running (either a localnet, or connected a live network), you should be able to run the following command to list the Protobuf services available (you can replace `localhost:9000` by the gRPC server endpoint of another node, which is configured under the `grpc.address` field inside `app.toml`):

```bash
grpcurl -plaintext localhost:9090 list
```

You should see a list of gRPC services, like `cosmos.bank.v1beta1.Query`. This is called reflection, which is a Protobuf endpoint returning a description of all available endpoints. Each of these represents a different Protobuf service, and each service exposes multiple RPC methods you can query against.

In the Cosmos SDK, we use [gogoprotobuf](https://github.com/gogo/protobuf) for code generation, and [grpc-go](https://github.com/grpc/grpc-go) for creating the gRPC server. Unfortunately, these two don't play well together, and more in-depth reflection (such as using grpcurl's `describe`) is not possible. See [this issue](https://github.com/grpc/grpc-go/issues/1873) for more info.

Instead, we need to manually pass the reference to relevant `.proto` files. For example:

```bash
grpcurl \
    -import-path ./proto \                              # Import these proto files too
    -import-path ./third_party/proto \                  # Import these proto files too
    -proto ./proto/cosmos/bank/v1beta1/query.proto \    # That's the proto file with the description of your service
    localhost:9090 \
    describe cosmos.bank.v1beta1.Query                  # Service we want to inspect
```

Once the Protobuf definitions are given, making a gRPC query is then straightforward, by calling the correct `Query` service RPC method, and by passing the request argument as data (`-d` flag):

```bash
grpcurl \
    -plaintext
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./proto/cosmos/bank/v1beta1/query.proto \
    -d '{"address":"$MY_VALIDATOR"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/AllBalances
```

The list of all available gRPC query endpoints is [coming soon](https://github.com/cosmos/cosmos-sdk/issues/7786).

### Query for historical state using gRPC

You may also query for historical data by passing some [gRPC metadata](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md) to the query: the `x-cosmos-block-height` metadata should contain the block to query. Using grpcurl as above, the command looks like:

```bash
grpcurl \
    -plaintext
    -import-path ./proto \
    -import-path ./third_party/proto \
    -proto ./proto/cosmos/bank/v1beta1/query.proto \
    -H "x-cosmos-block-height: 279256" \
    -d '{"address":"$MY_VALIDATOR"}' \
    localhost:9090 \
    cosmos.bank.v1beta1.Query/AllBalances
```

Assuming the state at that block has not yet been pruned by the node, this query should return a non-empty response.

### CosmJS

CosmJS documentation can be found at https://cosmos.github.io/cosmjs/. As of December 2020, CosmJS documentation is still work in progress.

## Using the REST Endpoints

As described in the [gRPC guide](../core/grpc_rest.md), all gRPC services on the Cosmos SDK are made available for more convenient REST-based queries through gRPC-gateway. The format of the URL path is based on the Protobuf service method's full-qualified name, but may contain small customizations so that final URLs look more idiomatic. For example, the REST endpoint for the `cosmos.bank.v1beta1.Query/AllBalances` method is `GET /cosmos/bank/v1beta1/balances/{address}`. Request arguments are passed as query parameters.

As a concrete example, the `curl` command to make balances request is:

```bash
curl \
    -X GET \
    -H "Content-Type: application/json" \
    http://localhost:1317/cosmos/bank/v1beta1/balances/$MY_VALIDATOR
```

Make sure to replace `localhost:1317` with the REST endpoint of your node, configured under the `api.address` field.

The list of all available REST endpoints is available as a Swagger specification file, it can be viewed at `localhost:1317/swagger`. Make sure that the `api.swagger` field is set to true in your `app.toml` file.

### Query for historical state using REST

Querying for historical state is done using the HTTP header `x-cosmos-block-height`. For example, a curl command would look like:

```bash
curl \
    -X GET \
    -H "Content-Type: application/json" \
    -H "x-cosmos-block-height: 279256"
    http://localhost:1317/cosmos/bank/v1beta1/balances/$MY_VALIDATOR
```

Assuming the state at that block has not yet been pruned by the node, this query should return a non-empty response.

## Next {hide}

Sending transactions using gRPC and REST requires some additional steps: generating the transaction, signing it, and finally broadcasting it. Read about [generating and signing transactions](TODO https://github.com/cosmos/cosmos-sdk/issues/7657). {hide}
