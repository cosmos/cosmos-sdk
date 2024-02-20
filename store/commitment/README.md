# State Commitment (SC)

The `commitment` package contains the state commitment (SC) implementation.
Specifically, it contains an IAVL v1 implementation of SC and the necessary types
and abstractions to support other SC backends, as well as supporting general integration
into store/v2, specifically the `RootStore` type.

A foremost design goal is that SC backends should be easily swappable, i.e. not
necessarily IAVL. To this end, the scope of SC has been reduced, it must only:

* Provide a stateful root app hash for height h resulting from applying a batch
  of key-value set/deletes to height h-1.
* Fulfill (though not necessarily provide) historical proofs for all heights < `h`.
* Provide an API for snapshot create/restore to fulfill state sync requests.

Notably, SC is not required to provide key iteration or value retrieval for either
queries or state machine execution, this now being the responsibility of state
storage.

An SC implementation may choose not to provide historical proofs past height `h - n`
(`n` can be 0) due to the time and space constraints, but since store/v2 defines
an API for historical proofs there should be at least one configuration of a
given SC backend which supports this.

## Benchmarks

See this [section](https://docs.google.com/document/d/1l6uXIjTPHOOWM5N4sUUmUfCZvePoa5SNfIEtmgvgQSU/edit#heading=h.7l0i621y5vgm) for specifics on SC benchmarks on various implementations.

## Usage
