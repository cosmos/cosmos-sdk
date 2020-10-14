# ADR 035: Rosetta API Support

## Context

Rosetta API, an open-source specification and set of tools developed by Coinbase, makes integration with blockchains simpler, faster, and more reliable by establishing a standard API for integrating blockchain applications.

By using a common interface that standardizes the process of how a user interacts with a blockchain, both the work of exchanges to integrate with new blockchains and also of the developers to build cross-blockchain applications such as block explorers, wallets and dApps is considerably reduced.

## Decision

We think that adding Rosetta API support to the Cosmos SDK will bring value to all the developers and Cosmos SDK based chains in the ecosystem.


## Architecture

The driving principles of the proposed design follow:

1. Developer-friendliness: it must be as riskless and painless as possible for client applications developers to extend network configurations to expose Rosetta API-compliant services.
2. Long term support: developers build applications on the various stable branches of Cosmos SDK. This proposal aims to provide support for all the release series supported by the Cosmos SDK team.
3. Cost-efficiency: backporting features from `master` to the various stable branches of Cosmos SDK is a cost that needs to be reduced.

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

The `Network` type holds network-specific properties (i.e. configuration values) and adapters. Pre-configured concrete types will be available for each Cosmos SDK release. Applications can also create their own custom types.

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

Cosmos SDK provides a base `Network` struct and a `NewNetwork` constructor that could serve as code example for client application developers and testing tool to be used in conjuction with `simd`.

```
// NewNetwork returns the default application configuration.
func NewNetwork(options Options) service.Network {
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

#### In-process Execution

Rosetta API service could run within the same execution process of the application; new configuration option and command line flag would be provided:

```
	if config.Rosetta.Enable {
     ....
            get contecxt, flags, etc
        	...
            
            h, err := service.New(
                service.Options{Port: config.Rosetta.Port},
                NewRosettaAPINetworkFromConfig(config),
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

#### Command Line Integration

Client application developers can write a new command to launch a Rosetta API server as a separate process too:

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
                service.Options{Port: config.Rosetta.Port},
                NewRosettaAPINetworkFromConfig(config),
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



### The external Library

Because apart from the Network struct and Adapter implemention there is a lot of code that would be shared across versions, we can provide a repo for the Rosetta Service that will hold only the Service, the Interfaces and the types. This includes all that is not specific to a single version. (Still to decide if we keep an external shared dependency or not.)


## Status

Accepted

## Consequences

### Positive

- Out-of-the-box Rosetta API support within Cosmos SDK.
- Blockchain interface standardisation

## References

- https://www.rosetta-api.org/
- https://github.com/tendermint/cosmos-rosetta-gateway
