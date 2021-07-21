<!--
order: false
parent:
  order: 2
-->

# Rosetta

This document provides instructions on how to use the Coinbase Rosetta API integration.

## Motivation and Design

For information about the motivation and design choices, refer to [ADR 035](../architecture/adr-035-rosetta-api-support.md).

## Usage

The Rosetta API server is a stand-alone server that connects to a node of a chain developed with the Cosmos SDK. 

To enable Rosetta API support, it's required to add the `RosettaCommand` to your application's root command file.
Find the following line within the root command file:

```
initRootCmd(rootCmd, encodingConfig)
```

After that line, add the following:

```
rootCmd.AddCommand(
  server.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Marshaler)
)
```


The `RosettaCommand` function builds the `rosetta` root command and is defined in the `server` package within Cosmos SDK.

Since we’ve updated the Cosmos SDK to work with the Rosetta API, updating the application's root command file is all you need to do.

To run Rosetta in your application CLI, use the following command:

```
appd rosetta --help
```

To test and run Rosetta API endpoints for applications that are running and exposed, use the following command:

```
appd rosetta
     --blockchain "your application name (ex: gaia)"
     --network "your chain identifier (ex: testnet-1)"
     --tendermint "tendermint endpoint (ex: localhost:26657)"
     --grpc "gRPC endpoint (ex: localhost:9090)"
     --addr "rosetta binding address (ex: :8080)"
```
