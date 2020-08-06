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
type Server struct {
    options
    router
}

type Options struct {}

func NewServer(opt Options,  rosettaAdapter Adapter) {
    return Server{opt, newRouter(rosettaAdapter)}    
}
```

Server is the main type that the client application needs to instantiate. It embeds an `Options` type and exposes a `Start()` method:

```
func (s Server) Start() error {
    ...
    s.ListenAndServe(s.options.endpoint, s.router)
    ...
}
```

An internal `router` type exposes the Rosetta API endpoints:

```
type router struct {
    muxRouter mux.router
    rosettaAdapter Adapter
}

func newRouter(adapter Adapter) {
    router := &Router{
        mux: mux.NewRouter(),
        rosettaAdapter: adapter,
    }

    router.mux.HandleFunc("/blocks/", router.blocksApi)
    ...
    ...

    return router
}
```

Each API endpoint has an associated handler, e.g.:

```
func (r router) blocksApi(w http.ResponseWriter, r *http.Request) {
    // We build the RosettaGetBlockRequest
    req := ....

    res := r.rosettaAdapter.GetBlocks(req) 
    ...

    // We write the response. 
}
``` 

The router just translates a rest request into the data struct that the interface of an adapter needs to run.
The RosettaAdapter interface looks like this:

```
type Adapter interface {
    GetBlock(req RosettaGetBlockRequest) RosettaGetBlockResponse
    ...
}
```

And we will provide different implementations of this `adapter` like for 0.38, for 0.39
and even a `CosmosHub` implementation that will call different versions of the hub.

An example of an adapter implementation looks like this:

```
type CosmosLaunchpad struct {}

func (CosmosLaunchpad) GetBlock(req RosettaGetBlockRequest) {
    // Here we would just convert the RosettaGetBlockRequest 
    // into the data we need to get the block information
    // from a node that uses Launchpad.

    res := client.Get(...)

    // The adapter is responsable to build the response in the type
    // specified by the Interface.

    return RosettaGetBlockResponse{...}
}
...

```

The adapters that come out-of-the-box can be easily overriden by application-specific implementations.

This way we offer developers the opportunity to embed
a new server instance in their applications. Developers always retain the possiblity to using the simple binary
that come with the Cosmos SDK.

## References

- https://www.rosetta-api.org/
