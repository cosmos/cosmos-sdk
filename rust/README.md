**WARNING: This is an API preview! Most code won't work or even type check properly!**

This is the single import, batteries-included crate for building applications with the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) in Rust.

## Core Concepts

* everything that runs code is an **account** with a unique [Address]
* the code that runs an account is called an **account handler**
* a **module** is a special type of account, whose **module handler** code gets instantiated once per app

## Creating an Account Handler

### Basic Structure

Follow these steps to create the basic structure for account handler:
1. Create a nested module (ex. `mod my_account_handler`) for the handler
2. Import this crate with `use ixc::*;` (_optional, but recommended_)
3. Add a handler struct to the nested `mod` block (ex. `pub struct MyAccountHandler`)
4. Annotate the struct with `#[derive(Resources)]`
5. Annotate the `mod` block with `#[ixc::account_handler(MyAccountHandler)]`

Here's an example:

```rust
#[ixc::account_handler(MyAsset)]
mod my_asset {
  use ixc::*;

  #[derive(Resources)]
  pub struct MyAsset {}
}
```

### Define the account's state

All the account's "resources" are managed by its handler struct.
Internal state is the primary "resource" that a handler interacts with.
(The other resources are references to other modules and accounts, which will be covered later.)
State is defined using the [`state_objects`] framework which defines the following basic state types:
[`Map`], [`Item`] and [`Set`], plus some other types extending these such as [`OrderedMap`],
[`OrderedSet`], [`Index`], [`UniqueIndex`] and [`UIntMap`].
See the [`state_objects`] documentation for more complete information.

The most basic type is [`Item`], which is a single value that can be read and written.
Here's an example of adding an item state resource to the handler:
```rust
#[derive(Resources)]
pub struct MyAsset {
    #[state(prefix=1)]
    pub owner: Item<Address>,
}
```

All state object resources should have the `#[state]` attribute.
The `prefix` attribute indicates the store key prefix and is optional, but recommended.

[`Map`] is a common type for any more complex state as it allows for multiple values to be stored and retrieved by a key. Here's an example of adding a map state resource to the handler:
```rust
#[derive(Resources)]
pub struct MyAsset {
    #[state(prefix=1)]
    pub owner: Item<Address>,
  
    #[state(prefix=2, key(account), value(amount))]
    pub balances: Map<Address, u128>,
}
```

Map state objects require `key` and `value` parameters in their `#[state]` attribute
in order to name the key and value fields in the map for querying by clients.

### Implement message handlers

Message handlers can be defined by attaching the `#[publish]` attribute to one of the
following targets:
* any inherent `impl` block for the handler struct, in which case all `pub` functions in that block will be treated as message handlers
* any `pub` function in an inherent `impl` block for the handler struct
* an `impl` block for the handler struct for any trait that has `#[account_api] attached to it

All message handler functions should immutably borrow the handler struct as the first argument (`&self`).
If they modify state, they should mutably borrow [`Context`] and
if they only read state, they should immutably borrow [`Context`].
Other arguments can be provided to the function signature as needed and the return
type should be [`Response`] parameterized with the return type of the function.
The supported argument types are defined by [`ixc_schema`] crate.
See that crate for more information.

Here's an example demonstrating all three methods:
```rust
#[publish]
impl MyAccountHandler {
    pub fn set_caller_value(&mut self, ctx: &Context, x: u64) -> Result<()> {
        self.my_map.set(ctx, ctx.caller(), x)?;
        Ok(())
    }
}

impl MyAccountHandler {
    #[publish]
    pub fn set_owner(&mut self, ctx: &Context, new_owner: Address) -> Result<()> {
        if ctx.caller() != self.owner(ctx)? {
            return Err("Unauthorized".into());
        }
        self.owner.set(ctx, new_owner)?;
        Ok(())
    }
}

#[account_api]
pub trait GetMyValue {
    fn get_my_value(&self, ctx: &Context) -> Response<u64>;
}

#[publish]
impl GetMyValue for MyAccountHandler {
    fn get_my_value(&self, ctx: &Context) -> Response<u64> {
        Ok(self.my_map.get(ctx, &ctx.caller())?)
    }
}
```
### Define an `on_create` method

One function in the handler struct can be defined as the "on create" method
by attaching the `#[on_create]` attribute to it.
This function will get called when the account is created and appropriate
arguments must be provided to it by the caller creating this account.
This method should return a [`Response<()>`].

Here's an example:
```rust
impl MyAsset {
    #[on_create]
    fn on_create(&mut self, ctx: &Context, initial_balance: u128) -> Result<()> {
        self.owner(ctx, ctx.caller())?;
        self.balances.set(ctx, &ctx.caller(), initial_balance)?;
        Ok(())
    }
}
```

### Emitting Events

Events can be emitted by adding [`EventBus`] parameters to method handler functions
where each [`EventBus`] is parameterized with an event type (usually a struct which
derives [`StructCodec`]).
Adding event buses for each type of event to each method ensures that the event API
is clearly defined in a handler's schema for external users.
(When client types are generated with `#[account_api]` or `#[module_api]` attributes,
the event bus parameters, however, are not included in the generated client types
so that callers don't need to worry about these.)

Here's an example of emitting an event:
```rust
#[publish]
impl MyAccountHandler {
    pub fn set_caller_value(&mut self, ctx: &Context, x: u64, events: &mut EventBus<SetValueVent>) -> Result<()> {
        self.my_map.set(ctx, ctx.caller(), x)?;
        events.emit(SetValueEvent { caller: ctx.caller.clone(), value: x });
        Ok(())
    }
}

#[derive(StructCodec)]
pub struct SetValueEvent {
    pub caller: Address,
    pub value: u64,
}
```

## Creating a Module Handler

The process for creating a module handler is very similar to creating an account handler.
The main difference is
that the [`module_handler`] attribute is used instead of the [`account_handler`] attribute
and that module handler may implement [`module_api`] traits in additional to [`account_api`]
traits.
A [`module_api`] trait is defined by attaching the `#[module_api]` attribute to a trait
and such traits can only be implemented by one module handler in the app.

Here's an example of a module handler:
```rust
#[ixc::module_handler(MyModuleHandler)]
mod my_module_handler {
    use ixc::*;
 
    #[derive(Resources)]
    pub struct MyModuleHandler {}

    #[module_api]
    pub trait MyModuleApi {
        fn my_module_fn(&self, ctx: &Context) -> Response<()>;
    }

    #[publish]
    impl MyModuleHandler {
        pub fn my_module_fn(&self, ctx: &Context) -> Response<()> {
            Ok(())
        }
    }
}
```

## Calling other accounts or modules

Any account may call any other account or module in the app by calling the generated
`Client` structs that are generated when traits are annotated with `#[account_api]` or `#[module_api]`.
Account API clients must be instantiated with the account's [`Address`] whereas
module API clients don't need to be parameterized with anything because
they must have a unique handler in the app.

Client or client factories should be defined as resources in the handler struct
in one of the following ways:
1. define a module API `Client` or `Option<Client>` as a resource in the handler struct
2. define an account API `ClientFactory` as a resource in the handler struct
3. define an account API `Client` or `Option<Client>` as a resource in the handler struct,
   annotated with `#[client]` and the address of the account

While these types can all be instantiated and called dynamically, it's better
to define them as explicit resources so that:
* framework tooling can ensure that API type definitions are consistent between different codebases
* the framework can ensure that all required modules are present in the app at startup

Here's an example of all these methods for defining client resources:
```rust
pub struct MyAccountHandler {
    pub my_module_client: MyModuleApiClient,
    pub optional_v2_client: Option<MyModuleApi2Client>,
    pub my_module_client_factory: MyModuleApiClientFactory,
    #[client("my_module_address")]
    pub my_module_client: Option<MyModuleApiClient>,
}
```

## Protobuf Compatibility

Most existing [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) applications use [Protocol Buffers](https://protobuf.dev) encoding.
So far, there has been no mention of it, and all of the examples here use plain old Rust
functions and structs so you may be wondering what's going on.

The [`ixc_schema`] crate defines a rich set of types that can be used in messages that aims to be
richer and more appropriate to this use case than the set of types provided by Protobuf.
Currently, many Cosmos SDK messages use Protobuf `string`s everywhere to represent other data
types.
Sometimes these are annotated with `cosmos_proto.scalar` to indicate which data type we actually mean.
In Rust, we encourage you to use the [`Address`] type for addresses and sized integer types, such as
`u128`, instead of strings or byte arrays.
For all of these types, there is a configurable mapping to Protobuf encoding that is described in more detail
in the [`ixc_schema`] crate documentation.
If the default mapping does not work, an alternate one may be available and can be annotated with optional
`#[proto]` and `#[schema]` attributes.
Ex:

```rust
#[derive(StructCodec)]
#[schema(name="cosmos.base.v1beta1.Coin")]
pub struct Coin {
    pub denom: String,
    #[proto(string, tag=2)] // tag could actually be inferred from field order, but shown for demonstration
    pub amount: u128,
}
```

Eventually, code generators may be implemented that take `.proto` files and generate this Rust code,
but for now we recommend following the Cosmos SDK's "expected keeper" pattern and just re-defining
types in Rust where needed.
Keeping client definitions close to where they're used avoids any of the pernicious issues
with versioning and dependencies that plague the Cosmos SDK in Golang.
Tooling is being developed to statically check the compatibility of types following the [`ixc_schema`] model
(see the `cosmosdk.io/schema/diff` package) so that type definitions from different Rust packages can be
compared for compatibility.

`#[proto]` and `#[schema]` annotations can also be used on arguments to handler functions
to configure the message names handlers correspond to.
(Note that even though the Cosmos SDK uses `service` definitions,
messages are actually routed by message name, not service method name.)
Ex:
```rust
#[module_api]
trait BankMsg {
   #[schema(name="cosmos.bank.v1beta1.MsgSend")]
   fn send(&self, ctx: &Context, 
           #[proto(string, msgv1_signer=true)] from: &Address,
           #[proto(string)] to: &Address,
           coins: &[Coin]) -> Response<()>;
}
```

## Exporting a Package of Handlers

The [`package_root!`] macro can be used to define a package of one or more handlers which can be
import as natively or as a virtual machine bundle (such as WASM):

```rust
package_root!(MyAccountHandler, MyModuleHandler);
```

## Testing

It is recommended that all account and module handlers write unit tests.
The [`ixc_testing`](https://docs.rs/ixc_testing) framework can be used for this purpose.

## Advanced Usage

### Splitting code across multiple files

The `#[ixc::account_handler]` and `#[ixc::module_handler]` attributes
work by searching for `#[publish]` and `#[on_create]` attributes in the same `mod` block.
To split code across multiple files, there are two options:
1. Reference the `#[account_api]` or `#[module_api]` traits by name in the `publish` field of the `#[ixc::account_handler]` or `#[ixc::module_handler]` attribute. Ex:
```rust
#[ixc::account_handler(MyAccountHandler, publish=[MyAccountApi])]
mod my_account_handler {
  // ...
}
```
2. Create account or module handlers in separate files and then reference them in the main handler struct
using [`ixc_core::handler::AccountMixin`] or [`ixc_core::handler::ModuleMixin`] types,
and then annotate these with `#[publish]`. 
Ex:
```rust
pub struct MyModuleHandler {
  #[publish]
  account_mixin: AccountMixin<NestedAccountHandler>,
  #[publish]
  module_mixin: ModuleMixin<NestedModuleHandler>,
}
```
`AccountMixin` and `ModuleMixin` implement the [`Deref`](core::ops::Deref) trait so that all methods
and types in those nested handlers are accessible through the mixin wrapper.

### Dynamically Routing Messages

All handler functions in [`#[account_api]`] and [`#[module_api]`] traits,
and inside [`#[publish]`] inherent `impl` blocks will have a corresponding message `struct` generated for them.
These structs can be used to dynamically invoke handlers using the [`Context::dynamic_invoke_module`] and
[`Context::dynamic_invoke_account`] methods.
Such structs can also be stored inside other structs and stored for later execution.

### Parallel Execution

**NOTE: this is a highly experimental design. During this API preview `parallel_safe` is enabled by default.**

The runtime executing account and module handler code written with this framework
may attempt to execute it in parallel with other code which may be attempting to
access the same state.
This parallel runtime will attempt to find a safe way to synchronize this state
access, usually by simulating transactions, tracking state reads and writes and
then scheduling things with appropriate ordering, checkpointing, and rollbacks.
The state that a handler is simulated against will likely be older than the
state that it is actually executed against, so there may be some differences in
behavior between simulation and execution.
Ideally, when a handler is simulated, it will have similar enough behavior to when it is
actually executed so that rollback, re-scheduling and re-execution are unnecessary.
If re-execution is necessary, the runtime may impose a penalty on the user
calling such handlers.
Generally, a handler will not need to be re-executed if it accesses the same
storage locations in simulation as during actual execution.
The values written to and read from those locations can vary, but if the
locations are the same, the runtime can ensure that the handler is only executed once.
Storage locations are identified by an account address and a state object's key.
We can ensure that such storage locations remain stable between simulation and execution
if the storage locations are derived _only_ from:
* message input,
* pure functions of message input, and
* older state that is guaranteed to be the same between simulation and execution

#### `parallel_safe` feature flag

The `state_objects` framework has a `parallel_safe` feature flag which uses lifetimes
and Rust's borrow checker to ensure that the above conditions are met at compile time.

In `parallel_safe` mode, `state_objects` types will have two lifetime parameters
called `'key` and `'value` and it is recommended that your handlers also declare
these lifetimes.
The `'key` lifetime represents things that will be stable between simulation and execution,
and are thus safe to use as state object keys to identify storage locations.
Only references with the `'key` lifetime should be used to derive state object keys.
References with `'value` lifetime should only be used as storage values as they may
change between simulation and execution.

Here's an example send method signature:
```rust
trait ParallelSafeSend {
    fn send<'key, 'value>(&self, ctx: &mut Context<'key>, to: &'key Address, denom: &'key str, amount: &'value u128) -> Response<()>;
}
```

This signature says that anything in the context as well as the address and denom
should be stable between simulation and execution, and thus can be used as keys.
The amount being sent, however, can change between simulation and execution.

A caller using this parallel safe `send` should then ensure that it only passes
references with the `'key` lifetime as arguments to the method (meaning derived
from message input, pure functions on that input, or previous block state).

#### Pure Functions

A pure function is a function that has no side effects and always returns the same
output given the same input.
Message handlers are pure if they have no `Context` parameter and thus have no
state access.
A pure function could be used to transform raw input parameters into some other
form while retaining the `'key` lifetime.
An example of this could be implementing a hash function as a pure function.
Then a storage key could be derived from the hash of some input parameters.

#### Stale Reads

In order to read a value from state that has the `'key` lifetime in `parallel_safe` mode,
the [`Map`] type has a `state_get` method which reads from some historical state
that is guaranteed to be the same between simulation and execution.
Regular calls to `get` will read from the latest state and return values with the `'value` lifetime.
However, calls to `state_get` will read from the historical state and return values with the `'key` lifetime which can then be used as keys for other state objects.

#### Lazy Writes

Lazy write operations is a technique to deal with resource contention during parallel execution.
Say we have some global balance (like a fee-pool) which is constantly being written to by
many transactions in a block.
So even if these transactions could otherwise run concurrently, they would all need
to lock around this fee-pool balance and actually need to run sequentially.

A lazy write operation allows us to work around this resource contention if we can ensure
that the order of write operations doesn't matter.
This is true if and only if the write operations are commutative.
This is the case when adding to a balance
as it can only fail when the underlying integer type saturates to its maximum value,
which is a fatal error condition anyway.

Because the framework has no way of knowing which write operations actually are
commutative, only privileged modules which are initialized by the application itself can
use lazy write operations.
The `state_objects` provides support for writing such modules using the [`UIntMap`] type
which has a `lazy_add` method.
If an unprivileged module tries to use `lazy_add`, the operation will occur synchronously.