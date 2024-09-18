**WARNING: Most of the code below does not work yet! Check back later!**

This is the single import, batteries-included crate for building applications with the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk) in Rust.

<!-- toc -->

- [Core Concepts](#core-concepts)
- [Creating an Account Handler](#creating-an-account-handler)
  * [Basic Structure](#basic-structure)
  * [Define the account's state](#define-the-accounts-state)
  * [Define the `OnCreate` Handler](#define-the-oncreate-handler)
  * [Implement message handlers](#implement-message-handlers)
  * [Emitting Events](#emitting-events)
- [Creating a Module Handler](#creating-a-module-handler)
- [Calling other accounts or modules](#calling-other-accounts-or-modules)
- [Exporting a Package of Handlers](#exporting-a-package-of-handlers)
- [Testing](#testing)

<!-- tocstop -->

## Core Concepts

* everything that runs code is an **account** with a unique [Address]
* the code that runs an account is called an **account handler**
* a **module** is a special type of account, whose **module handler** code gets instantiated once per app

## Creating an Account Handler

### Basic Structure

Follow these steps to create the basic structure for account handler:
1. Create a nested module (ex. `mod my_account_handler`) for the handler
2. Import this crate with `use interchain_sdk::*;` (_optional, but recommended_)
3. Add a handler struct to the nested `mod` block (ex. `pub struct MyAccountHandler`)
4. Annotate the struct with `#[derive(Resources)]`
5. Annotate the `mod` block with `#[interchain_sdk::account_handler(MyAccountHandler)]`

Here's an example:

```rust
#[interchain_sdk::account_handler(MyAsset)]
mod my_asset {
  use interchain_sdk::*;

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
[`OrderedSet`], [`Index`], [`UniqueIndex`] and [`UInt128Map`].
See the [`state_objects`] documentation for more complete information.

The most basic type is [`Item`], which is a single value that can be read and written.
Here's an example of adding an item state resource to the handler:
```rust
#[derive(Resources)]
pub struct MyAsset {
    #[item(prefix=1)]
    pub owner: Item<Address>,
}
```

[`Item`] fields should have the `#[item]` attribute, and the type of the field should be `Item<T>`.
The `prefix` attribute is optional, but recommended and is used to specify the store prefix for the item.

[`Map`] is a common type for any more complex state as it allows for multiple values to be stored and retrieved by a key. Here's an example of adding a map state resource to the handler:
```rust
#[derive(Resources)]
pub struct MyAsset {
    #[item(prefix=1)]
    pub owner: Item<Address>,
  
    #[map(prefix=2, key(account), value(amount))]
    pub balances: Map<Address, u128>,
}
```

[`Map`] fields should have the `#[map]` attribute, and the type of the field should be `Map<K, V>`.
The `key` and `value` attributes are required
and specify the field names of the key and value fields in the map which are necessary for indexing maps for querying.

### Define the `OnCreate` Handler

Every account handler must implement the [`OnCreate`] trait,
which defines the behavior when an account is created.
This is where you can set any initial state of the account.
The [`OnCreate`] implementation must define an `InitMessage`,
which is the message that is passed to the account when it is created.
That struct must derive [`StructCodec`] which allows
it to be serialized and deserialized.

Here's an example:
```rust
#[derive(StructCodec)]
pub struct MyAccountCreateMsg {
    pub initial_balance: u128,
}

impl OnCreate for MyAsset {
    fn on_create(&mut self, ctx: &Context, msg: &CreateMsg) -> Result<()> {
        self.owner(ctx, ctx.caller())?;
        self.balances.set(ctx, &ctx.caller(), msg.init_value)?;
        Ok(())
    }
}
```

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
The supported argument types are defined by [`interchain_schema`] crate.
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
#[interchain_sdk::module_handler(MyModuleHandler)]
mod my_module_handler {
    use interchain_sdk::*;
 
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

## Exporting a Package of Handlers

The [`package_root!`] macro can be used to define a package of one or more handlers which can be
import as natively or as a virtual machine bundle (such as WASM):

```rust
package_root!(MyAccountHandler, MyModuleHandler);
```

## Testing

It is recommended that all account and module handlers write unit tests.
The `interchain_core_testing` framework can be used for this purpose.