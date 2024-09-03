# RFC 003: Cross Language Account/Module Execution Model

## Changelog

* 2024-08-09: Reworked initial draft (previous work was in https://github.com/cosmos/cosmos-sdk/pull/15410)

## Background

The Cosmos SDK has historically been a Golang only framework for building blockchain applications.
However, discussions about supporting additional programming languages and virtual machine environments
have been underway since early 2023. Recently, we have identified the following key target user groups:
Recently we have identified the following key target user groups:
1. projects that want to primarily target a single programming language and virtual machine environment besides Golang but who still want to use Cosmos SDK internals for consensus and storage
2. projects that want to integrate multiple programming languages and virtual machine environments into an integrated application

While these two user groups may have substantially different needs,
the goals of the second group are more challenging to support and require a more clearly specified unified design.

This RFC primarily attempts to address the needs of the second group.
However, in doing so, it also intends to address the needs of the first group as we will likely build many of the components needed for this group by building an integrated cross-language, cross-VM framework.
Those needs of the first group which are not satisfactorily addressed by the cross-language framework should be addressed in separate RFCs.

Prior work on cross-language support in the SDK includes:
- [RFC 003: Language-independent Module Semantics & ABI](https://github.com/cosmos/cosmos-sdk/pull/15410): an earlier, significantly different and unmerged version of this RFC.
- [RFC 002: Zero Copy Encoding](./rfc-002-zero-copy-encoding.md): a zero-copy encoding specification for ProtoBuf messages, which was partly implemented and may or may not still be relevant to the current work.

Also, this design largely builds on the existing `x/accounts` module and extends that paradigm to environments beyond just Golang.
That design was specified in [RFC 004: Accounts](./rfc-004-accounts.md).

## Proposal

We propose a conceptual and formal model for defining **accounts** and **modules** which can interoperate with each other through **messages** in a cross-language, cross-VM environment.

We start with the conceptual definition of core concepts from the perspective of a developer
trying to write code for a module or account. 
The formal details of how these concepts are represented in a specific coding environment may vary significantly,
however, the essence should remain more or less the same in most coding environments. 

This specification is intentionally kept minimal as it is always easier to add features than to remove them.
Where possible, other layers of the system should be specified in a complementary, modular way in separate specifications.

### Account

An **account** is defined as having:
* a unique **address**
* an **account handler** which is some code which can process **messages** and send **messages** to other **accounts**

### Address

An **address** is defined as a variable-length byte array of up to 63 bytes
so that an address can be represented as a 64-byte array with the first byte indicating the length of the address.

### Message

A **message** is defined as a tuple of:
* a **message name** 
* and **message data**

A **message name** is an ASCII string of up to 127 characters
so that it can be represented as a 128-byte array with the first byte indicating the length of the string. 
**Message names** can only contain letters, numbers and the special characters `:`, `_`, `/`, and `.`.

**Message data** will be defined in more detail later.

### Account Handler

The code that implements an account's message handling is known as the **account handler**. The handler receives a **message request** and can return some **message response** or an error.

The handler for a specific message within an **account handler** is known as a **message handler**.

### Message Request

A **message request** contains:
* the **address** of the **account** (its own address)
* the **address** of the account sending the message (the **caller**), which will be empty if the message is a query
* the **message name**
* the **message data**
* a 32-byte **state token**
* a 32-byte **context token**
* a `uint64` **gas limit**

**Message requests** can also be prepared by **account handlers** to send **messages** to other accounts.

### Modules and Modules Messages

There is a special class of **message**s known as **module messages**,
where the caller should omit the address of the receiving account.
The routing framework can look up the address of the receiving account based on the message name of a **module message**.

Accounts which define handlers for **module messages** are known as **modules**.

**Module messages** are distinguished from other messages because their message name must start with the `module:` prefix.

The special kind of account handler which handles **module messages** is known as a **module handler**.
A **module** is thus an instance of a **module handler** with a specific address 
in the same way that an account is an instance of an account handler.
In addition to an address, **modules** also have a human-readable **module name**.

More details on **modules** and **module messages** will be given later.

### Account Handler and Message Metadata

Every **account handler** is expected to provide metadata which provides:
* a list of the **message names** it defines **message handlers** and for each of these, its:
  * **volatility** (described below)
  * optional additional bytes, which are not standardized at this level
* **state config** bytes which are sent to the **state handler** (described below) but are otherwise opaque
* some optional additional bytes, which are not standardized at this level

### Account Lifecycle

**Accounts** can be created, destroyed and migrated to new **account handlers**.

**Account handlers** can define message handlers for the following special message name's:
* `on_create`: called when an account is created with message data containing arbitrary initialization data.
* `on_migrate`: called when an account is migrated to a new code handler. Such handlers receive structured message data specifying the old code handler so that the account can perform migration operations or return an error if migration is not possible.

### Hypervisor and Virtual Machines

Formally, a coding environment where **account handlers** are run is known as a **virtual machine**.
These **virtual machine**s may or may not be sandboxed virtual machines in the traditional sense.
For instance, the existing Golang SDK module environment (currently specified by `cosmossdk.io/core`), will
be known as the "native Golang" virtual machine.
For consistency, however,
we refer to these all as **virtual machines** because from the perspective of the cross-language framework,
they must implement the same interface.

The special module which manages **virtual machines** and **accounts** is known as the **hypervisor**.

Each **virtual machine** that is loaded by the **hypervisor** will get a unique **machine id** string.
Each **account handler** that a **virtual machine** can load is referenced by a unique **handler id** string.

There are two forms of **handler ids**:
* **module handlers** which take the form `module:<module_config_name>`
* **account handlers** which take the form `<machine_id>:<machine_handler_id>`, where `machine_handler_id` is a unique string scoped to the **virtual machine**

Each **virtual machine** must expose a list of all the **module handlers** it can run,
and the **hypervisor** will ensure that the **module handlers** are unique across all **virtual machines**.

Each **virtual machine** is expected to expose a method which takes a **handler id**
and returns a reference to an **account handler**
which can be used to run **messages**.
**Virtual machines** will also receive an `invoke` function
so that their **account handlers** can send messages to other **accounts**.
**Virtual machines** must also implement a method to return the metadata for each **account handler** by **handler id**.

### State and Volatility

Accounts generally also have some mutable state, but within this specification,
state is mostly expected to be handled by some special state module defined by separate specifications.
The few main concepts of **state handler**, **state token**, **state config** and **volatility** are defined here.

The **state handler** is a system component which the hypervisor has a reference to,
and which is responsible for managing the state of all accounts.
It only exposes the following methods to the hypervisor:
- `create(account address, state config)`: creates a new account state with the specified address and **state config**.
- `migrate(account address, new state config)`: migrates the account state with the specified address to a new state config
- `destroy(account address)`: destroys the account state with the specified address

**State config** are optional bytes that each account handler's metadata can define which get passed to the **state handler** when an account is created.
These bytes can be used by the **state handler** to determine what type of state and commitment store the **account** needs.

A **state token** is an opaque array of 32-bytes that is passed in each message request.
The hypervisor has no knowledge of what this token represents or how it is created,
but it is expected that modules that mange state do understand this token and use it to manage all state changes
in consistent transactions.
All side effects regarding state, events, etc. are expected to coordinate around the usage of this token.
It is possible that state modules expose methods for creating new **state tokens**
for nesting transactions.

**Volatility** describes a message handler's behavior with respect to state and side effects.
It is an enum value that can have one of the following values:
* `volatile`: the handler can have side effects and send `volatile`, `radonly` or `pure` messages to other accounts. Such handlers are expected to both read and write state.
* `readonly`: the handler cannot cause effects side effects and can only send `readonly` or `pure` messages to other accounts. Such handlers are expected to only read state.
* `pure`: the handler cannot cause any side effects and can only call other pure handlers. Such handlers are expected to neither read nor write state.

The hypervisor will enforce **volatility** rules when routing messages to account handlers.
Caller addresses are always passed to `volatile` methods,
they are not required when calling `readonly` methods but will be passed when available,
and they are not passed at all to `pure` methods.

### Management of Account Lifecycle with the Hypervisor

In order to manage **accounts** and their mapping to **account handlers**, the **hypervisor** contains stateful mappings for:
* **account address** to **handler id**
* **module name** to module **account address** and **module config**
* **message name** to **account address** for **module messages**

The **hypervisor** as a first-class module itself handles the following special **module messages** to manage account
creation, destruction, and migration:
* `create(handler_id, init_data) -> address`: creates a new account in the specified code environment with the specified handler id and returns the address of the new account. The `on_create` message is called if it is implemented by the  account. Addresses are generated deterministically by the hypervisor with a configurable algorithm which will allow public key accounts to get predictable addresses.
* `destroy(address)`: deletes the account with the specified address. `destroy` can only be called by the account itself.
* `migrate(address, new_handler_id)`: migrates the account with the specified address to the new account handler. The `on_migrate` message must be implemented by the new code and must not return an error for migration to succeed. `migrate` can only be called by the account itself.
* `force_migrate(address, new_handler_id, init_data)`: this can be used when no `on_migrate` handler can perform a proper migration to the new account handler. In this case, the old account state will be destroyed, and `on_create` will be called on the new code. This is a destructive operation and should be used with caution.

The **hypervisor** will call the **state handler**'s `create`, `migrate`,
and `destroy` methods as needed when accounts are created, migrated, or destroyed.

### Module Lifecycle & Module Messages

For legacy purposes, **modules** have specific lifecycles and **module messages** have special semantics.
A **module handler** cannot be loaded with the `create` message,
but must be loaded by an external call to the hypervisor
which includes the **module name** and **module config** bytes.
The existing `cosmos.app.v1alpha1.Config` can be used for this purpose if desired.

**Module messages** also allow the definition of pre- and post-handlers.
These are special message handlers that can only be defined in **module handlers**
and must be prefixed by the `module:pre:` or `module:post:` prefixes
When modules are loaded in the hypervisor, a composite message handler will be composed using all the defined
pre- and post-handlers for a given message name in the loaded module set.
By default, the ordering will be done alphabetically by module name.

### Authorization and Delegated Execution

When a message handler creates a message request, it can pass any address as the caller address,
but it must pass the same **context token** that it received in its message request.
The hypervisor will use the **context token** to verify the "real" caller address.
Every nested message call will receive a new non-forgeable **context token** so that virtual machines
and their account handlers cannot arbitrarily fool the hypervisor about the real caller address.

By default, the hypervisor will only allow the real caller to act as the caller.

There are use cases, however, for delegated authorization of messages or even for modules which can execute
a message on behalf of any account.
To support these, the hypervisor will accept an **authorization middleware** parameter which checks
whether a given real caller account (verified by the hypervisor) is authorized to act as a different caller
account for a given message request.

### Message Data and Packet Specification

To facilitate efficient cross-language and cross-VM message passing, the precise layout of **message packets** is important
as it reduces the need for serialization and deserialization in the core hypervisor and virtual machine layers.

We start by defining a **message packet** as a 64kb (65,536 bytes) array which is aligned to a 64kb boundary.
For most message handlers, this single packet should be large enough to contain a full **message request**,
including all **message data** as well as message return data.
In cases where the packet size is too small, additional buffers can be referenced from within the **message packet**.

More details on the specific layout of **message packets** will be specified in a future update to this RFC
or a separate RFC.
For now, we specify that within a 64kb **message packet**,
at least 56kb will be available for **message data** and message responses.

## Abandoned Ideas (Optional)

## Decision

Based on internal discussions, we have decided to move forward with this design. 

## Consequences (optional)

### Backwards Compatibility

It is intended that existing SDK modules built using `cosmossdk.io/core` and
account handlers built with `cosmossdk.io/x/accounts` can be integrated into this system with zero or minimal changes.

### Positive

This design will allow native SDK modules to be built using other languages such as Rust and Zig, and
for modules to be executed in different virtual machine environments such as Wasm and the EVM.
It also extends the concept of a module to first-class accounts in the style of the existing `x/accounts` module
and EVM contracts.

### Negative

### Neutral

Similar to other message passing designs,
the raw performance invoking a message handler will be slower than a golang method call as in the existing keeper paradigm.

However, this design does nothing to preclude the continued existence of golang native keeper passing, and it is likely
that we can find performance optimizations in other areas to mitigate any performance loss.
In addition, a cross-language, cross-VM is simply not possible without some overhead.


### References

- [Abandoned RFC 003: Language-independent Module Semantics & ABI](https://github.com/cosmos/cosmos-sdk/pull/15410)
- [RFC 002: Zero Copy Encoding](./rfc-002-zero-copy-encoding.md) 
- [RFC 004: Accounts](./rfc-004-accounts.md)

## Discussion

This specification does not cover many important parts of a complete system such as the encoding of message data,
storage, events, transaction execution, or interaction with consensus environments.
It is the intention of this specification to specify the minimum necessary for this layer in a modular layer.
The full framework should be composed of a set of independent, minimally defined layers that together
form a "standard" execution environment, but that at the same time can be replaced and recomposed by
different applications with different needs.

The basic set of standards necessary to provide a coherent framework includes:
* message encoding and naming, including compatibility with the existing protobuf-based message encoding
* storage
* events
* authorization middleware
