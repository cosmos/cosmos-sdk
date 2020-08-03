# ADR 26: Rosetta API Support

## Changelog

- August 3rd, 2020: Initial Draft.

## Status

- Approved, pending fine tuning the details of implementation.

## Context

We think it'd be greatly valuable to application developers to have the Cosmos SDK
provide them with out-of-the-box Rosetta API support.
According to [the project's website](https://www.rosetta-api.org/), Rosetta API is an open
standard designed to simplify blockchain deployment and interaction. The latest specifications are
available at [this URL](https://www.rosetta-api.org/docs/Reference.html).

We want to achieve the following objectives:

- Support multiple versions of Cosmos SDK.
- Support Cosmos Hub.
- Implement querying of historical data sets.

## Decision

We intend to develop a library that could be extended and used by application
developers to integrate an in-process Rosetta API-compliant server with the
application main binaries. We also intend to provide a standalone gateway server
program that supports a Cosmos SDK's minimum feature set. Such program could
run alongside the client applications main binaries.

### Implementation

```
type Server struct {}
func NewServer(opt Options)

type Options struct {}
```

This struct once started will listen to the port specified by the options and will
expose the Rosetta API.

Internally we will have an interface that will abstract the different calls that Rosetta supports.

Example:

```
type RosettaDataAPI interface {
    GetBlock(req RosettaGetBlockRequest) RosettaGetBlockResponse
    ...
}
```

And we will provide different implementations of this `adapter` like for 0.38, for 0.39
and even a `CosmosHub` implementation that will call different versions of the hub.

This way we offer the possibility to offer developers the opportunity to instantiate
a new server in their applications or just using the binary that can be build from the
repository.

## References

- https://www.rosetta-api.org/
