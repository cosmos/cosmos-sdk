# ADR 26: Rosetta API

## Changelog

- August 3rd, 2020: Initial Draft.

## Status

- Approved, pending fine tuning the details of implementation.

## Context

There is the need to provide an option to expose the Cosmos SDK applications to use
the Rosetta API. This is an open Standard designed to simplify and integrate different
blockchains.

Our use case need to be able to provide:

- Support for the different versions of the Cosmos SDK.
- Support for different versions of the Cosmos Hub.
- Support to be able to provide an abstraction in order to provide querying data, let's say
we have a TxHash which we don't know if it is in cosmos hub 1, 2, 3. But queryng the Data
API will return us the transaction without the need to query 3 versions.

## Decision

In order to provide this use cases we have decided that the best approach will be to 
use an external repository that will hold the libraries that can be integrated in
different applications if they want to embed it into their own application and a
main.go that will provide a binary that will run as a standalone.

## References

- https://www.rosetta-api.org/
