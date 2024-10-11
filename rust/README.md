**WARNING: This is an API preview! Expect major bugs, glaring omissions, and breaking changes!**

This is the single import, batteries-included crate for building applications with the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) in Rust.

## Core Concepts

* everything that runs code is an **account** with a unique [AccountID]
* the code that runs an account is called an **handler**

## Creating a Handler

### Basic Structure

Follow these steps to create the basic structure for account handler:
1. Create a nested module (ex. `mod my_handler`) for the handler
2. Import this crate with `use ixc::*;` (_optional, but recommended_)
3. Add a handler struct to the nested `mod` block (ex. `pub struct MyHandler`)
4. Annotate the struct with `#[derive(Resources)]`
5. Annotate the `mod` block with `#[ixc::handler(MyHandler)]`

Here's an example:

```rust
#[ixc::handler(MyAsset)]
mod my_asset {
  use ixc::*;

  #[derive(Resources)]
  pub struct MyHandler {}
}
```

### Define the account's state

All the account's "resources" are managed by its handler struct.
Internal state is the primary "resource" that a handler interacts with.
(The other resources are references to other modules and accounts, which will be covered later.)
State is defined using the [`state_objects`] framework which defines types for storing and retrieving state.
See the [`state_objects`] documentation for more complete information.

The most basic state object type is [`Item`], which is a single value that can be read and written.
Here's an example of adding an item state resource to the handler:
```rust
#[derive(Resources)]
pub struct MyHandler {
    #[state(prefix=1)]
    pub owner: Item<AccountID>,
}
```

All state object resources should have the `#[state]` attribute.
The `prefix` attribute indicates the store key prefix and is optional, but recommended.

[`Map`] is a common type for any more complex state as it allows for multiple values to be stored and retrieved by a key. Here's an example of adding a map state resource to the handler:
```rust
#[derive(Resources)]
pub struct MyHandler {
    #[state(prefix=1)]
    pub owner: Item<AccountID>,
  
    #[state(prefix=2, key(account), value(amount))]
    pub my_map: Map<AccountID, u64>,
}
```

Map state objects require `key` and `value` parameters in their `#[state]` attribute
in order to name the key and value fields in the map for querying by clients.

### Implement message handlers

Message handlers can be defined by attaching the `#[publish]` attribute to one of the
following targets:
* any inherent `impl` block for the handler struct, in which case all `pub` functions in that block will be treated as message handlers
* any `pub` function in an inherent `impl` block for the handler struct
* an `impl` block for the handler struct for any trait that has `#[handler_api] attached to it

All message handler functions should immutably borrow the handler struct as the first argument (`&self`).
If they modify state, they should mutably borrow [`Context`] and
if they only read state, they should immutably borrow [`Context`].
Other arguments can be provided to the function signature as needed and the return
type should be [`Result`] parameterized with the return type of the function.
The supported argument types are defined by [`ixc_schema`] crate.
See that crate for more information.

Here's an example demonstrating all three methods:
```rust
#[publish]
impl MyHandler {
    pub fn set_caller_value(&self, ctx: &mut Context, x: u64) -> Result<()> {
        self.my_map.set(ctx, ctx.caller(), x)?;
        Ok(())
    }
}

impl MyHandler {
    #[publish]
    pub fn set_owner(&self, ctx: &mut Context, new_owner: Address) -> Result<()> {
        if ctx.caller() != self.owner(ctx)? {
            return Err("Unauthorized".into());
        }
        self.owner.set(ctx, new_owner)?;
        Ok(())
    }
}

#[handler_api]
pub trait GetMyValue {
    fn get_my_value(&self, ctx: &Context) -> Result<u64>;
}

#[publish]
impl GetMyValue for MyHandler {
    fn get_my_value(&self, ctx: &Context) -> Result<u64> {
        Ok(self.my_map.get(ctx, &ctx.caller())?)
    }
}
```
### Define an `on_create` method

One function in the handler struct must be defined as the "on create" method
by attaching the `#[on_create]` attribute to it.
This function will get called when the account is created and appropriate
arguments must be provided to it by the caller creating this account.
This method should return a [`Result<()>`].

Here's an example:
```rust
impl MyHandler {
    #[on_create]
    fn on_create(&mut self, ctx: &Context, initial_value: u64) -> Result<()> {
        self.owner(ctx, ctx.caller())?;
        self.my_map.set(ctx, &ctx.caller(), initial_value)?;
        Ok(())
    }
}
```

### Emitting Events

Events can be emitted by adding [`EventBus`] parameters to method handler functions
where each [`EventBus`] is parameterized with an event type (usually a struct which
derives [`SchemaValue`]).
Adding event buses for each type of event to each method ensures that the event API
is clearly defined in a handler's schema for external users.
(When client types are generated, however,
the event bus parameters are not included in the generated client types
so that callers don't need to worry about these.)

Here's an example of emitting an event:
```rust
#[publish]
impl MyAccountHandler {
    pub fn set_caller_value(&mut self, ctx: &Context, x: u64, evt_bus: &mut EventBus<SetValueVent>) -> Result<()> {
        self.my_map.set(ctx, ctx.caller(), x)?;
        evt_bus.emit(SetValueEvent { caller: ctx.caller.clone(), value: x });
        Ok(())
    }
}

#[derive(SchemaValue)]
pub struct SetValueEvent {
    pub caller: AccountID,
    pub value: u64,
}
```

## Calling other accounts

Any account may call any other account or module in the app by calling the client structs
that are generated for handlers and `#[handler_api]` traits.

Clients can be defined as resources in the handler struct using the `#[client]`
attribute and the `AccountID` as an integer (NOTE: more robust ways of setting this are planned).

While clients can be instantiated and called dynamically, it's better
to define them as explicit resources so that:
* framework tooling can ensure that API type definitions are consistent between different codebases
* the framework can ensure that the required accounts are present in the app at startup

Here's an example of defining a client resource for `#[handler_api]` trait.
In this case we must cast that trait dynamically to `Service` to get its client type
(ex. `<dyn MyTrait as Service>::Client`):
```rust
pub struct MyHandler {
    #[client(123456789)]
    pub get_my_value_client: <dyn GetMyValue as Service>::Client,
}
```

### Dynamically Routing Messages

All handler functions in [`#[handler_api]`] traits,
and inside [`#[publish]`] inherent `impl` blocks will have a corresponding message `struct` generated for them.
These structs can be used to dynamically invoke handlers using [`ixc_core::low_level::dynamic_invoke`].
Such structs can also be placed inside other structs and stored for later execution.

### Creating new accounts

Accounts can be created in tests or by other accounts using the [`create_account`] function.
This function must be parameterized with the handler type and the struct generated by its `#[on_create]` method, ex: `create_account::<MyHandler>(&mut ctx, MyHandlerCreate { initial_value: 42 })`.

### Error Handling

All functions should return the [`Result`] type with an error message if the function fails.
Error messages can be created using the `error!` macro, ex: `Err(error!("Invalid input"))`,
or the `bail!` or `ensure!` macros (similar to as in the `anyhow` crate).
[`Result`] can also be parameterized with custom error codes.
See the `examples/` directory for more examples on usage.

## Testing

The [`ixc_testing`](https://docs.rs/ixc_testing) framework can be used for writing unit
and integration tests for handlers and has support for mocking.
See its documentation for more information as well as the `examples/` directory for more examples on usage.

[//]: # (## Advanced Usage)

[//]: # (### Splitting code across multiple files)

[//]: # ()
[//]: # (The `#[ixc::account_handler]` and `#[ixc::module_handler]` attributes)

[//]: # (work by searching for `#[publish]` and `#[on_create]` attributes in the same `mod` block.)

[//]: # (To split code across multiple files, there are two options:)

[//]: # (1. Reference the `#[account_api]` or `#[module_api]` traits by name in the `publish` field of the `#[ixc::account_handler]` or `#[ixc::module_handler]` attribute. Ex:)

[//]: # (```rust)

[//]: # (#[ixc::account_handler&#40;MyAccountHandler, publish=[MyAccountApi]&#41;])

[//]: # (mod my_account_handler {)

[//]: # (  // ...)

[//]: # (})

[//]: # (```)

[//]: # (2. Create account or module handlers in separate files and then reference them in the main handler struct)

[//]: # (using [`ixc_core::handler::AccountMixin`] or [`ixc_core::handler::ModuleMixin`] types,)

[//]: # (and then annotate these with `#[publish]`. )

[//]: # (Ex:)

[//]: # (```rust)

[//]: # (pub struct MyModuleHandler {)

[//]: # (  #[publish])

[//]: # (  account_mixin: AccountMixin<NestedAccountHandler>,)

[//]: # (  #[publish])

[//]: # (  module_mixin: ModuleMixin<NestedModuleHandler>,)

[//]: # (})

[//]: # (```)

[//]: # (`AccountMixin` and `ModuleMixin` implement the [`Deref`]&#40;core::ops::Deref&#41; trait so that all methods)

[//]: # (and types in those nested handlers are accessible through the mixin wrapper.)

[//]: # (### Parallel Execution)

[//]: # ()
[//]: # (**NOTE: this is a highly experimental design. During this API preview `parallel_safe` is enabled by default.**)

[//]: # ()
[//]: # (The runtime executing account and module handler code written with this framework)

[//]: # (may attempt to execute it in parallel with other code which may be attempting to)

[//]: # (access the same state.)

[//]: # (This parallel runtime will attempt to find a safe way to synchronize this state)

[//]: # (access, usually by simulating transactions, tracking state reads and writes and)

[//]: # (then scheduling things with appropriate ordering, checkpointing, and rollbacks.)

[//]: # (The state that a handler is simulated against will likely be older than the)

[//]: # (state that it is actually executed against, so there may be some differences in)

[//]: # (behavior between simulation and execution.)

[//]: # (Ideally, when a handler is simulated, it will have similar enough behavior to when it is)

[//]: # (actually executed so that rollback, re-scheduling and re-execution are unnecessary.)

[//]: # (If re-execution is necessary, the runtime may impose a penalty on the user)

[//]: # (calling such handlers.)

[//]: # (Generally, a handler will not need to be re-executed if it accesses the same)

[//]: # (storage locations in simulation as during actual execution.)

[//]: # (The values written to and read from those locations can vary, but if the)

[//]: # (locations are the same, the runtime can ensure that the handler is only executed once.)

[//]: # (Storage locations are identified by an account address and a state object's key.)

[//]: # (We can ensure that such storage locations remain stable between simulation and execution)

[//]: # (if the storage locations are derived _only_ from:)

[//]: # (* message input,)

[//]: # (* pure functions of message input, and)

[//]: # (* older state that is guaranteed to be the same between simulation and execution)

[//]: # ()
[//]: # (#### `parallel_safe` feature flag)

[//]: # ()
[//]: # (The `state_objects` framework has a `parallel_safe` feature flag which uses lifetimes)

[//]: # (and Rust's borrow checker to ensure that the above conditions are met at compile time.)

[//]: # ()
[//]: # (In `parallel_safe` mode, `state_objects` types will have two lifetime parameters)

[//]: # (called `'key` and `'value` and it is recommended that your handlers also declare)

[//]: # (these lifetimes.)

[//]: # (The `'key` lifetime represents things that will be stable between simulation and execution,)

[//]: # (and are thus safe to use as state object keys to identify storage locations.)

[//]: # (Only references with the `'key` lifetime should be used to derive state object keys.)

[//]: # (References with `'value` lifetime should only be used as storage values as they may)

[//]: # (change between simulation and execution.)

[//]: # ()
[//]: # (Here's an example send method signature:)

[//]: # (```rust)

[//]: # (trait ParallelSafeSend {)

[//]: # (    fn send<'key, 'value>&#40;&self, ctx: &mut Context<'key>, to: &'key Address, denom: &'key str, amount: &'value u128&#41; -> Result<&#40;&#41;>;)

[//]: # (})

[//]: # (```)

[//]: # ()
[//]: # (This signature says that anything in the context as well as the address and denom)

[//]: # (should be stable between simulation and execution, and thus can be used as keys.)

[//]: # (The amount being sent, however, can change between simulation and execution.)

[//]: # ()
[//]: # (A caller using this parallel safe `send` should then ensure that it only passes)

[//]: # (references with the `'key` lifetime as arguments to the method &#40;meaning derived)

[//]: # (from message input, pure functions on that input, or previous block state&#41;.)

[//]: # ()
[//]: # (#### Pure Functions)

[//]: # ()
[//]: # (A pure function is a function that has no side effects and always returns the same)

[//]: # (output given the same input.)

[//]: # (Message handlers are pure if they have no `Context` parameter and thus have no)

[//]: # (state access.)

[//]: # (A pure function could be used to transform raw input parameters into some other)

[//]: # (form while retaining the `'key` lifetime.)

[//]: # (An example of this could be implementing a hash function as a pure function.)

[//]: # (Then a storage key could be derived from the hash of some input parameters.)

[//]: # ()
[//]: # (#### Stale Reads)

[//]: # ()
[//]: # (In order to read a value from state that has the `'key` lifetime in `parallel_safe` mode,)

[//]: # (the [`Map`] type has a `state_get` method which reads from some historical state)

[//]: # (that is guaranteed to be the same between simulation and execution.)

[//]: # (Regular calls to `get` will read from the latest state and return values with the `'value` lifetime.)

[//]: # (However, calls to `state_get` will read from the historical state and return values with the `'key` lifetime which can then be used as keys for other state objects.)

[//]: # ()
[//]: # (#### Lazy Writes)

[//]: # ()
[//]: # (Lazy write operations is a technique to deal with resource contention during parallel execution.)

[//]: # (Say we have some global balance &#40;like a fee-pool&#41; which is constantly being written to by)

[//]: # (many transactions in a block.)

[//]: # (So even if these transactions could otherwise run concurrently, they would all need)

[//]: # (to lock around this fee-pool balance and actually need to run sequentially.)

[//]: # ()
[//]: # (A lazy write operation allows us to work around this resource contention if we can ensure)

[//]: # (that the order of write operations doesn't matter.)

[//]: # (This is true if and only if the write operations are commutative.)

[//]: # (This is the case when adding to a balance)

[//]: # (as it can only fail when the underlying integer type saturates to its maximum value,)

[//]: # (which is a fatal error condition anyway.)

[//]: # ()
[//]: # (Because the framework has no way of knowing which write operations actually are)

[//]: # (commutative, only privileged modules which are initialized by the application itself can)

[//]: # (use lazy write operations.)

[//]: # (The `state_objects` provides support for writing such modules using the [`UIntMap`] type)

[//]: # (which has a `lazy_add` method.)

[//]: # (If an unprivileged module tries to use `lazy_add`, the operation will occur synchronously.)