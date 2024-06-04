# Indexer Base

The indexer base module is designed to provide a stable, zero-dependency base layer for the built-in indexer functionality. Packages that integrate with the indexer should feel free to depend on this package without fear of any external dependencies being pulled in.

The basic types for specifying index sources, targets and decoders are provided here along with a basic engine that ties these together. A package wishing to be an indexing source could accept an instance of `Engine` directly to be compatible with indexing. A package wishing to be a decoder can use the `Entity` and `Table` types. A package defining an indexing target should implement the `Indexer` interface.