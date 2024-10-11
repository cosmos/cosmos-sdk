**WARNING: This is an API preview! Expect major bugs, glaring omissions, and breaking changes!**

This is the single import, batteries-included crate for building applications with the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) in Rust.

## Core Concepts

* everything that runs code is an **account** with a unique [AccountID]
* the code that runs an account is called a **handler**

## Getting Started

Follow these steps to create the basic structure for account handler:
1. Create a nested module (ex. `mod my_handler`) for the handler
2. Import this crate with `use ixc::*;` (_optional, but recommended_)
3. Add a handler struct to the nested `mod` block (ex. `pub struct MyHandler`)
4. Annotate the struct with `#[derive(Resources)]`
5. Annotate the `mod` block with `#[ixc::handler(MyHandler)]`
6. Define an `#[on_create]` method for the handler struct

Here's an example:

```rust
#[ixc::handler(MyHandler)]
mod my_handler {
  use ixc::*;

  #[derive(Resources)]
  pub struct MyHandler {}
    
  impl MyHandler {
    #[on_create]
    fn create(&mut self, ctx: &Context) -> Result<()> { Ok(()) }
  }
}
```

The `#[on_create]` method will be called when the account is created and must return a [`Result<()>`].
It can take additional arguments as needed.

## Managing State

All the account's "resources" are managed by its handler struct.
Internal state is the primary "resource" that a handler interacts with.
(The other resources are references to other modules and accounts, which will be covered later.)
State is defined using the [`state_objects`] framework which defines types for storing and retrieving state.
See the [`state_objects`] documentation for more complete information.

### Item

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

### Map

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

## Other State Objects

See the [`state_objects`] documentation for more information on other state object types.
In particular, the [`Accumulator`] and [`AccumulatorMap`] types are useful whenever
any sort of balance or supply tracking is needed.

## Publishing message handlers

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
The supported argument types are that implement [`ixc_schema::SchemaValue`].
See the [`ixc_schema`] crate for more information.

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

## Emitting Events

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

NOTE: for now, events don't do anything and are discarded. This will be fixed in an upcoming release.

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

### Sending dynamic messages

All handler functions in [`#[handler_api]`] traits,
and inside [`#[publish]`] inherent `impl` blocks will have a corresponding message `struct` generated for them.
These structs can be used to dynamically invoke handlers using [`ixc_core::low_level::dynamic_invoke`].
Such structs can also be placed inside other structs and stored for later execution.

## Creating new accounts

Accounts can be created in tests or by other accounts using the [`create_account`] function.
This function must be parameterized with the handler type and the struct generated by its `#[on_create]` method, ex: `create_account::<MyHandler>(&mut ctx, MyHandlerCreate { initial_value: 42 })`.

## Error Handling

All functions should return the [`Result`] type with an error message if the function fails.
Error messages can be created using the `error!` macro, ex: `Err(error!("Invalid input"))`,
or the `bail!` or `ensure!` macros (similar to as in the `anyhow` crate).
[`Result`] can also be parameterized with custom error codes.
See the `examples/` directory for more examples on usage.

## Testing

The [`ixc_testing`](https://docs.rs/ixc_testing) framework can be used for writing unit
and integration tests for handlers and has support for mocking.
See its documentation for more information as well as the `examples/` directory for more examples on usage.
