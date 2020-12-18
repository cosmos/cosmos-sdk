# ADR 035: Rosetta API Support

## Authors

- Jonathan Gimeno (@jgimeno)
- David Grierson (@senormonito)
- Alessio Treglia (@alessio)

## Context

[Rosetta API](https://www.rosetta-api.org/) is an open-source specification and set of tools developed by Coinbase to 
standardise blockchain interactions.

Through the use of a standard API for integrating blockchain applications it will

* Be easier for a user to interact with a given blockchain
* Allow exchanges to integrate new blockchains quickly and easily
* Enable application developers to build cross-blockchain applications such as block explorers, wallets and dApps at 
  considerably lower cost and effort.

## Decision

It is clear that adding Rosetta API support to the Cosmos SDK will bring value to all the developers and 
Cosmos SDK based chains in the ecosystem. How it is implemented is key.

The driving principles of the proposed design are:

1. **Extensibility:** it must be as riskless and painless as possible for application developers to set-up network 
   configurations to expose Rosetta API-compliant services.
2. **Long term support:** This proposal aims to provide support for all the supported Cosmos SDK release series.
3. **Cost-efficiency:** Backporting changes to Rosetta API specifications from `master` to the various stable 
   branches of Cosmos SDK is a cost that needs to be reduced.

We will achieve these delivering on these principles by the following:

1. There will be an external repo called [cosmos-rosetta-gateway](https://github.com/tendermint/cosmos-rosetta-gateway) 
   for the implementation of the core Rosetta API features, particularly:
   a. The types and interfaces. This separates design from implementation detail.
   b. Some core implementations: specifically, the `Service` functionality as this is independent of the Cosmos SDK version.
2. Due to differences between the Cosmos release series, each series will have its own specific API implementations of `Network` struct and `Adapter` interface.
3. There will be two options for starting an API service in applications:
   a. API shares the application process
   b. API-specific process.


## Architecture

### The External Repo

As section will describe the proposed external library, including the service implementation, plus the defined types and interfaces.

#### Service

`Service` is a simple `struct` that is started and listens to the port specified in the options. This is meant to be used across all the Cosmos SDK versions that are actively supported.

The constructor follows:

`func New(options Options, network Network) (*Service, error)`

#### Types

`Service` accepts an `Options` `struct` that holds service configuration values, such as the port the service would be listening to:

```golang
type Options struct {
    ListenAddress string
}
```

The `Network` type holds network-specific properties (i.e. configuration values) and adapters. Pre-configured concrete types will be available for each Cosmos SDK release. Applications can also create their own custom types.

```golang
type Network struct {
	Properties rosetta.NetworkProperties
	Adapter    rosetta.Adapter
}
```

A `NetworkProperties` `struct` comprises basic values that are required by a Rosetta API `Service`:

```golang
type NetworkProperties struct {
	// Mandatory properties
	Blockchain          string
	Network             string
	SupportedOperations []string
}
```

Rosetta API services use `Blockchain` and `Network` as identifiers, e.g. the developers of _gaia_, the application that powers the Cosmos Hub, may want to set those to `Cosmos Hub` and `cosmos-hub-3` respectively.

`SupportedOperations` contains the transaction types that are supported by the library. At the present time, 
only `cosmos-sdk/MsgSend` is supported in Launchpad. Additional operations will be added in due time.

For Launchpad we will map the amino type name to the operation supported, in Stargate we will use the protoc one.

#### Interfaces

Every SDK version uses a different format to connect (rpc, gRpc, etc), we have abstracted this in what is called the 
Adapter. This is an interface that defines the methods an adapter implementation must provide in order to be used 
in the `Network` interface.

Each Cosmos SDK release series will have their own Adapter and NodeClient implementations.
Developers can implement their own custom adapters as required.

```golang
type Adapter interface {
	DataAPI
	ConstructionAPI
}

type DataAPI interface {
	server.NetworkAPIServicer
	server.AccountAPIServicer
	server.MempoolAPIServicer
	server.BlockAPIServicer
	server.ConstructionAPIServicer
}

type ConstructionAPI interface {
	server.ConstructionAPIServicer
}
```

For further information about the `Servicer` interfaces, please refer to the [Coinbase's rosetta-sdk-go's documentation](https://pkg.go.dev/github.com/coinbase/rosetta-sdk-go@v0.5.9/server).

Example in pseudo-code of an Adapter interface:

```golang
type SomeAdapter struct {
	cosmosClient     client
	tendermintClient client
}

func NewSomeAdapter(nodeClient NodeClient) rosetta.Adapter {
	return &SomeAdapter{client: nodeClient}
}

func (s SomeAdapter) NetworkStatus(ctx context.Context, request *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	resp := s.client.CallStatus()
	// ... Parse status Response
	// build NetworkStatusResponse
	return networkStatusResp, nil
}

func (s SomeAdapter) AccountBalance(ctx context.Context, request *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	resp := s.client.Account()
	// ... Parse cosmos specific account response
	// build AccountBalanceResponse
	return AccountBalanceResponse, nil
}

// And we repeat for all the methods defined in the interface.
```

The NodeClient represents a client that talks with the node, normally this is different on every cosmos-sdk. We already provide
an implementation for it.

```golang
// NodeClient defines the interface
// a client must implement in order to
// interact with the client application.
type NodeClient interface {
   // Balances fetches the balance of the given address
   // if height is not nil, then the balance will be displayed
   // at the provided height, otherwise last block balance will be returned
   Balances(ctx context.Context, addr string, height *int64) ([]*types.Amount, error)
   // BlockByHashAlt gets a block and its transaction at the provided height
   BlockByHash(ctx context.Context, hash string) (BlockResponse, error)
   // BlockByHeightAlt gets a block given its height, if height is nil then last block is returned
   BlockByHeight(ctx context.Context, height *int64) (BlockResponse, error)
   // BlockTransactionsByHash gets the block, parent block and transactions
   // given the block hash.
   BlockTransactionsByHash(ctx context.Context, hash string) (BlockTransactionsResponse, error)
   // BlockTransactionsByHash gets the block, parent block and transactions
   // given the block hash.
   BlockTransactionsByHeight(ctx context.Context, height *int64) (BlockTransactionsResponse, error)
   // GetTx gets a transaction given its hash
   GetTx(ctx context.Context, hash string) (*types.Transaction, error)
   // GetUnconfirmedTx gets an unconfirmed Tx given its hash
   // NOTE(fdymylja): NOT IMPLEMENTED YET!
   GetUnconfirmedTx(ctx context.Context, hash string) (*types.Transaction, error)
   // Mempool returns the list of the current non confirmed transactions
   Mempool(ctx context.Context) ([]*types.TransactionIdentifier, error)
   // Peers gets the peers currently connected to the node
   Peers(ctx context.Context) ([]*types.Peer, error)
   // Status returns the node status, such as sync data, version etc
   Status(ctx context.Context) (*types.SyncStatus, error)

	PostTx(txBytes []byte) (res *types.TransactionIdentifier, meta map[string]interface{}, err error)
	SignedTx(ctx context.Context, txBytes []byte, sigs []*types.Signature) (signedTxBytes []byte, err error)
	TxOperationsAndSignersAccountIdentifiers(signed bool, hexBytes []byte) (ops []*types.Operation, signers []*types.AccountIdentifier, err error)
	ConstructionMetadataFromOptions(ctx context.Context, options map[string]interface{}) (meta map[string]interface{}, err error)
	ConstructionPayload(ctx context.Context, req *types.ConstructionPayloadsRequest) (resp *types.ConstructionPayloadsResponse, err error)
	PreprocessOperationsToOptions(ctx context.Context, req *types.ConstructionPreprocessRequest) (options map[string]interface{}, err error)
	SupportedOperations() []string
}
```

A developer that wants to support their custom messages in order to be parsed and constructed by the Rosetta API must implement
the following interface.

```golang
// Msg interface is the interface that Cosmos SDK messages should implement if they want to
// be supported by the Rosetta service.
type Msg interface {
   ToOperations(withStatus bool, hasError bool) []*types.Operation
   FromOperations(ops []*types.Operation) (sdk.Msg, error)
}
```

To get an idea how a message implementing this interface looks you can take a look at MsgSend from bank module.


### 2. Cosmos SDK Implementation

As described, each Cosmos SDK release series will have version specific implementations of `Network` and `Adapter`, as 
well as a `NewNetwork` constructor.

Due to separation of interface and implementation, application developers have the option to override as needed, 
using this code as reference.

```golang
// NewNetwork builds a rosetta gateway network
func NewNetwork(networkIdentifier *types.NetworkIdentifier, adapter crg.Adapter, client NodeClient) service.Network {
   return service.Network{
      Properties: crg.NetworkProperties{
         Blockchain:          networkIdentifier.Blockchain,
         Network:             networkIdentifier.Network,
         AddrPrefix:          sdk.GetConfig().GetBech32AccountAddrPrefix(), // since we're inside cosmos sdk the config is supposed to be sealed
         SupportedOperations: client.SupportedOperations(),
      },
      Adapter: adapter,
   }
}
```

#### Extending Custom Messages
We assumed that every sdk.Msg will have a 1 to many relation with Rosetta operations, meaning that every sdk.Message that implements
the rosetta.Msg will be parsed and be able to be used via Rosetta API.

Once this message is registered in the Interface registry it will be automatically part of the Rosetta API engine.

### 3. API service invocation

As stated at the start, application developers will have two methods for invocation of the Rosetta API service:

1. Shared process for both application and API
2. Standalone API service

#### Shared Process (Only Stargate)

Rosetta API service could run within the same execution process as the application. New configuration option and 
command line flags would be provided to support this:

```golang
	if config.Rosetta.Enable {
     ....
            get contecxt, flags, etc
        	...
            
            h, err := service.New(
                service.Options{ListenAddress: config.Rosetta.ListenAddress},
                rosetta.NewNetwork(cdc, options),
            )
            if err != nil {
            }
            
            ...
            
            go func() {
			if err := h.Start(config); err != nil {
				errCh <- err
			}
		    }()
    }

```

#### Separate API service

Client application developers can write a new command to launch a Rosetta API server as a separate process too:

```golang
func RosettaCommand(cdc *codec.Codec) *cobra.Command {

    ...
    cmd := &cobra.Command{
    	Use:   "rosetta",
        ....
        
		RunE: func(cmd *cobra.Command, args []string) error {
            ....
            get contecxt, flags, etc
        	...
            
            h, err := service.New(
                  service.Options{Endpoint: endpoint},
                  rosetta.NewNetwork(cdc, options),
            )
            if err != nil {
            	return err
            }
            
            ...
            
            h.Start()
        }
    }
    ...

}
```

## Status

Proposed

## Consequences

### Positive

- Out-of-the-box Rosetta API support within Cosmos SDK.
- Blockchain interface standardisation

## References

- https://www.rosetta-api.org/
- https://github.com/tendermint/cosmos-rosetta-gateway
