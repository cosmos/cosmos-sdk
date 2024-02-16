## Example Module

```rust
#[module(services=[MsgServer, QueryServer])]
struct Bank {
    store: StoreClient,
    events: EventsClient,
}

impl MsgServer for Bank {
}

impl QueryServer for Bank {
}
```

The `module` attribute implements the `Module` trait.

`Module` trait has:
* descriptor
* init function
* route function

## Example Module Bundle

```rust
#![module_bundle(modules=[Bank])]
```

This would generate a static variables with:
* a module bundle containing an instance of all modules which implements a router
* proto descriptors
* init entry points