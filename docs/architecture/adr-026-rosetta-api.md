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

In order to provide this use cases we have decided that the best approach will be to 
use an external repository that will hold the libraries that can be integrated in
different applications if they want to embed it into their own application and a
main.go that will provide a binary that will run as a standalone.

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
