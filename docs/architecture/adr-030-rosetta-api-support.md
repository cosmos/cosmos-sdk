# ADR 033: Rosetta API support

## Context

Rosetta API, an open-source specification and set of tools developed by Coinbase, makes integration with blockchains simpler, faster, and more reliable by establishing a standard API for integrating blockchain applications.

By using a common interface that standardizes the process of how a user interacts with a blockchain, both the work of exchanges to integrate with new blockchains and also of the developers to build cross-blockchain applications such as block explorers, wallets and dApps is considerably reduced.

We think that providing Rosetta support to the Cosmos SDK will add value to all the developers and Cosmos SDK based chains in the ecosystem.

## Decision


The service is structured in a way that:

1. It becomes easy to inject different implementations for different types of SDK. For this abstraction we have used the term Adapter.
2. Due to the nature of versioning that has been done with Cosmos SDK so far it becomes very difficult to have different SDK versions included in the same repo, for that we created a generic shared code that includes the Service and the Adapter interface.
3. It is easy to inject and instantiate wherever is needed in the different applications.


## Architecture


### Service

`Service` is a simple `struct` that is started and listens to the port specified in the options. This is meant to be used across all the Cosmos SDK versions that are actively supported.

The constructor follows:

`func New(options Options, network Network) (*Service, error)`

It accepts an `Options` `struct` that holds service configuration values, such as the port the service would be listening to:

```
type Options struct {
	Port uint32
}
```

The `Network` type holds network-specific properties (i.e. configuration values) and adapters. Ready-to-use Cosmos SDK release series-specific base `Network` concrete types will be made available. Applications can embed such types in their own custom types.

```
type Network struct {
	Properties rosetta.NetworkProperties
	Adapter    rosetta.Adapter
}
````

A `NetworkProperties` `struct` comprises basic values that are required by a Rosetta API `Service` to run:

```
type NetworkProperties struct {
	// Mandatory properties
	Blockchain          string
	Network             string
	AddrPrefix          string
	SupportedOperations []string
}
```

Rosetta API services use Blockchain and Network as identifiers, e.g. the developers of gaia, the application that powers the Cosmos Hub, may want to set those to Cosmos Hub and cosmos-hub-3 respectively.

`AddrPrefix` contains the network-specific address prefix. Cosmos SDK base implementations will default to `cosmos`, client applications are instructed that this should be changed according to their network configuration.

`SupportedOperations` contains the transaction types that are supported by the library. At the present time, only `Transfer` is supported.

`Network` holds an `Adapter` reference too. Adapter implementations may vary across different Cosmos SDK release series:

And adapter must satisfy the interface:

```
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

````

### Cosmos SDK Integration

The best way to do it would be that every Cosmos SDK version holds the implementation of the Network struct, this can be an example located in the Cosmos SDK Launchpad version.

```
// NewLaunchpadNetwork returns a configured network to work in a Launchpad version.
func NewLaunchpadNetwork(options Options) service.Network {
	cosmosClient := cosmos.NewClient(fmt.Sprintf("http://%s", options.CosmosEndpoint))
	tendermintClient := tendermint.NewClient(fmt.Sprintf("http://%s", options.TendermintEndpoint))

	return service.Network{
		Properties: rosetta.NetworkProperties{
			Blockchain:          options.Blockchain,
			Network:             options.Network,
			SupportedOperations: []string{OperationTransfer},
		},
		Adapter: newAdapter(
			cosmosClient,
			tendermintClient,
			properties{
				Blockchain:  options.Blockchain,
				Network:     options.Network,
				AddrPrefix:  options.AddrPrefix,
				OfflineMode: options.OfflineMode,
			},
		),
	}
}
```

As we see we hold not only the instantiation of a Launchpad specific network but the adapter implementation too. As the way to talk to different rpc endpoint change between different versions (Launchpad, Stargate, etc), makes sense to keep it in their respective branches.

In this way we could have two options (yet to decide) that we can expose rosetta in the cosmos SDK.

#### Embedding it in the Start command.

This will look more like the approach the API in cosmos works: [link](https://github.com/cosmos/cosmos-sdk/blob/ca7b31dd5d23b10a919de1a60bbac90e89c5bf78/server/start.go#L248).

So we can add something like:

```
	if config.Rosetta.Enable {
     ....
            get contecxt, flags, etc
        	...
            
            h, err := service.New(
                service.Options{Port: uint32(*flagPort)},
                launchpad.NewLaunchpadNetwork(launchpad.Options{
                    CosmosEndpoint:     *flagAppRPC,
                    TendermintEndpoint: *flagTendermintRPC,
                    Blockchain:         *flagBlockchain,
                    Network:            *flagNetworkID,
                    AddrPrefix:         *flagAddrPrefix,
                    OfflineMode:        *flagOfflineMode,
                }),
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

#### The cobra command approach

This means providing a new command in order to run the service, an example can look something like:

```
func NewRosettaServiceCmd() *cobra.Command {

    ...
    cmd := &cobra.Command{
    	Use:   "vote [proposal-id] [option]",
        ....
        
		RunE: func(cmd *cobra.Command, args []string) error {
            ....
            get contecxt, flags, etc
        	...
            
            h, err := service.New(
                service.Options{Port: uint32(*flagPort)},
                launchpad.NewLaunchpadNetwork(launchpad.Options{
                    CosmosEndpoint:     *flagAppRPC,
                    TendermintEndpoint: *flagTendermintRPC,
                    Blockchain:         *flagBlockchain,
                    Network:            *flagNetworkID,
                    AddrPrefix:         *flagAddrPrefix,
                    OfflineMode:        *flagOfflineMode,
                }),
            )
            if err != nil {
            }
            
            ...
            
            h.Start()
        }
    }
    ...

}
```

This is closer to the Standalone approach. Right now in the current implementation we have a main.go that can be run as a Standalone, this probably will be removed once we migrate the code from Tendermint repo an embed it in Cosmos SDK.


### The external Library

Because apart from the Network struct and Adapter implemention there is a lot of code that would be shared across versions, we can provide a repo for the Rosetta Service that will hold only the Service, the Interfaces and the types. This includes all that is not specific to a single version. (Still to decide if we keep an external shared dependency or not.)


## Status

Accepted

## Consequences

### Positive

- Provide out-of-the-box Rosetta interface just by using the Cosmos SDK.
- Contribute to the standarization of blockchain interfaces.

## References

- https://www.rosetta-api.org/
- https://github.com/tendermint/cosmos-rosetta-gateway
