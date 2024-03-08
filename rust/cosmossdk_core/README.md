
## Example Module

```rust
// the module attribute specifies the services that the module provides
// so that these can be used to generate the descriptor and router code
#[module(services=[
    MsgServer,
    QueryServer,
    InternalServer,
    BeginBlocker,
    Handler<SomeMsg>,
    EventHook<SomeEvent>,
])]
struct Bank {
    // the module defines its dependencies as fields
    // and the module attribute uses this to generate
    // the descriptor and initialization code
    store: StoreClient,
    events: EventClient,
    
    // the module can also define its configuration message as an input
    config: BankModuleConfig,
}

impl MsgServer for Bank {
    // the context is mutable for messages
    fn send(&self, ctx: &mut zeropb::Context, req: &MsgSend) -> zeropb::Result<MsgSendResponse> {
        // ...
    }
}

impl QueryServer for Bank {
    // the context is immutable for queries
    fn balance(&self, ctx: &zeropb::Context, req: &QueryBalanceRequest) -> zeropb::Result<QueryBalanceResponse> {
        // ...
    }
}

// modules can also implement internal services which use a caller ID for authentication
// rather than trying to extract the signer from the message - this will be much simpler
// to route
impl InternalServer for Bank {
    fn mint(&self, ctx: &mut zeropb::Context, caller: CallerID, req: &InternalMintRequest) -> zeropb::Result<InternalMintResponse> {
        // ...
    }
}

// begin and end block can be special services defined on modules
impl BeginBlocker for Bank {
    fn begin_block(&self, ctx: &mut zeropb::Context, req: &BeginBlockRequest) -> zeropb::Result<BeginBlockResponse> {
        // ...
    }
}

impl Handler<SomeMsg> for Bank {
    fn handler(&self, ctx: &mut zeropb::Context, msg: &SomeMsg) -> zeropb::Result<()> {
        // ...
    }
}

#[derive(ZeroPB)]
struct SomeMsg {
  sender: zeropb::Str
}

impl EventHook<SomeEvent> for Bank {
    fn on_event(&self, ctx: &mut zeropb::Context, event: &SomeEvent) -> zeropb::Result<()> {
        // ...
    }
}
```

## Example Module Bundle

```rust
#![module_bundle(modules=[Bank, Staking, Governance])]
```

This code would define all functions needed for the module bundle and expose them via WASM, FFI or some other mechanism
depending on the target environment.

## ZeroInit

ZeroInit would be an extension of ZeroPB for making the initialization process of a module bundle essentially zero cost
and allowing module bundles to be compiled without even having an allocator. This would allow WASM and other VM bundle
sizes to be very small and performant.

If we look at the `Bank` module struct defined above it basically contains three fields: `store`, `events`, and `config`.
`store` and `events` are service clients which just contain a pointer to the connection function and a service ID which are
essentially two `usize` integers. `config` is the encoded configuration message which when encoded with ZeroPB is just any other
ZeroPB data. Well the host knows the addresses of these pointers and service IDs, so the host can actually encode this
whole configuration struct using ZeroPB.

The `module_bundle` attribute itself would them generate a `ModuleBundle` struct which contains each of the module structs, ex:
```rust
struct ModuleBundle {
    bank: zeropb::Option<Bank>,
    staking: zeropb::Option<Staking>,
    governance: zeropb::Option<Governance>,
}
```

Each of these module structs would be wrapped in an `Option` because we don't know if the host wants to initialize them or not -
that is dependent on the app configuration.

With this layout, the host can actually encode the whole `ModuleBundle` struct using ZeroPB and then pass it to `init` and
`init` can return the same pointer which can be called by invoke directly. There is essentially zero initialization needed
The host can cache the proto and module descriptors and even the value of the module bundle itself, so if this were
used for WASM, the WASM module could be loaded and initialized with a single very quick call to the host.

## Module Bundle ABI

A group of modules in a single compilation unit is called a module bundle.

Module bundles expose the following functions (shown with Rust syntax for clarity):
* `fn file_descriptor_set() -> Vec<u8>` - returns the proto file descriptor set of the module bundle, can be gzip or bzip2 compressed and the magic bytes of the compression format can be used to detect it.
* `fn default_encoding() -> u32` (optional) - returns the default encoding of the module bundle as defined by the `Encoding` enum. If this function is not implemented, the default encoding is assumed to be ZeroPB.
* `fn module_descriptors() -> Vec<u8>`- returns a byte array containing `ModuleBundleDescriptor` encoded with the default encoding.
* `fn init(init_data: Vec<u8) -> usize` - initializes the module bundle with the given init data and returns a handle to the router. The init data is the data that the module bundle needs to initialize itself. The return value is the handle to the router, which is an integer that is used to route messages to the module bundle.
* `fn invoke(router_handle: usize, method: usize, context: usize, caller: usize, p1: *mut (), p2: *mut ()) -> usize` - routes a message handler.
  Its exact implementation may vary a bit depending on the encoding used, but this signature assumes zeropb and fixed sized 64kb buffers. The `router_handle` parameter is the router returned by `init`, the `method` parameter is the method index, the `context` parameter is the context pointer, `caller` is the optional ID of the caller for authentication when methods use it, and the `p1` and `p2` parameters are generally the request and response points, but their use may depend on the method being invoked. The return value is the response code but its use may depend on the method being invoked.
* `fn deinit(usize)` -> () - deinitializes the module bundle and frees the resources associated with it. The parameter is the router handle returned by `init`.
* `alloc` and `free` as necessary for the encodings used - for zeropb, these functions always return single 64kb buffers.

The host must define a single import function, `invoke_host` which has a very similar signature to invoke above and is used by modules to call methods routed by the host. The module bundle may be able to route some messages to other modules in the bundle without using the host, but only in the case when the interaction can be properly authorized (which is easy with queries, possible with proposed "internal" services, and more complex with messages where authentication is done via the signer field).
