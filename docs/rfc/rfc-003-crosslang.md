# RFC 003: Account/Module Execution Model

## Changelog

* 2024-08-09: Reworked initial draft (previous work was in https://github.com/cosmos/cosmos-sdk/pull/15410)

## Background

The Cosmos SDK has historically been a Golang only framework for building blockchain applications.
However, discussions about supporting additional programming languages and virtual machine environments
have been underway since early 2023.
Recently we have identified the following key target user groups:
1. projects that want to primarily target a single programming language and virtual machine environment besides Golang but who still want to use Cosmos SDK internals for consensus and storage
2. projects that want to integrate multiple programming languages and virtual machine environments into an integrated application

While these two user groups may have substantially different needs,
the goals of the second group are more challenging to support and require a more clearly specified unified design.

This RFC primarily attempts to address the needs of the second group.
However, in doing so, it also intends to address the needs of the first group as we will likely build many of the components needed for this group by building an integrated cross-language, cross-VM framework.
Those needs of the first group which are not satisfactorily addressed by the cross-language framework should be addressed in separate RFCs.

## Proposal

We propose a conceptual and formal model for defining **accounts** and **modules** which can interoperate with each other through **messages** in a cross-language, cross-VM environment.

We start with the conceptual definition of core concepts from the perspective of a developer
trying to write code for a module or account. 
The formal details of how these concepts are represented in a specific coding environment may vary significantly,
however, the essence should remain more or less the same in most coding environments. 

### Account

An **account** is defined as having:
* a unique **address**
* an **account handler** which is some code which can process **messages** and send **messages** to other **accounts**
* some optional **account config** bytes

### Address

An **address** is defined as a variable-length byte array of up to 255 bytes
so that an address can be represented as a 256-byte array with the first byte indicating the length of the address.
TODO: a 255-byte is probably way longer than needed and we could likely shorten this to 63 bytes without any negative impact.

### Message

A **message** is defined as a tuple of:
* a **message name** 
* and **message data**

A **message name** is an ASCII string of up to 255 characters so that it can be represented as a 256-byte array with the first byte indicating the length of the string. 
**Message names** can only contain letters, numbers and the special characters `:`, `_`, `/`, and `.`.

**Message data** will be defined in more detail later.

### Account Handler

The code that implements an account's message handling is known as the **account handler**.

When a **message** is sent to an **account handler**, it will receive:
* the **address** of the **account** (its own address)
* the **address** of the account sending the message (the caller), which will be empty if the message is a query
* the **message name**
* the **message data**
* a `uint64` **gas limit**
* optional **account config** bytes

The handler can then execute some code and return a response or an error. Details on message responses and errors will be described later.

To send a **message** to another **account**, the caller must specify:
* the **message name**
* the **message data**
* a **gas limit**
* optionally, the **address** of the **account** to send the message to

The handler for a specific message within an **account handler** is known as a **message handler**.

### Modules and Modules Messages

There is a special class of **message**s known as **module messages**,
where the caller should omit the **address** of the receiving **account**.
The routing framework can look up the **address** of the receiving **account** based on the **message name** of a **module message**.

**Accounts** which define handlers for **module messages** are known as **modules**.

**Module messages** are distinguished from other messages because their **message name** must start with the `module:` prefix.

The special kind of **account handler** which handles **module messages** is known as a **module handler**.
A **module** is thus instance an instance of a **module handler** with a specific **address** 
in the same way that an **account** is an instance of an **account handler**.
In addition to a bytes **address**, **modules** also have a human-readable **module name**.

More details on **modules** and **module messages** will be given later.

### Account Handler and Message Metadata

Every **account handler** is expected to provide metadata which at a minimum provides:
* a list of the **message names** it defines **message handlers** for, and
* the **volatility** of each message handler, described below

Metadata for **account handlers** and **message handlers** can also contain additional bytes
which for now are not standardized at this level.

### State and Volatility

**Account**s generally also have some mutable state, but within this specification, state is expected to be handled by some special state **module** which is defined by separate specifications.

However, the metadata for each message handler in an **account handler** does specify its **volatility**,
which describes a message handler's behavior with respect to state and side effects.
**Volatility** is an enum value that can have one of the following values:
* `volatile`: the handler can have side effects and send `volatile`, `radonly` or `pure` messages to other accounts. Such handlers are expected to both read and write state.
* `readonly`: the handler cannot cause effects side effects and can only send `readonly` or `pure` messages to other accounts. Such handlers are expected to only read state.
* `pure`: the handler cannot cause any side effects and can only call other pure handlers. Such handlers are expected to neither read nor write state.

### Account Lifecycle

**Accounts** can be created, destroyed and migrated to new **account handlers**.

**Account handlers** can define code for the following special **message name**'s:
* `on_create`: called when an account is created with **message data** containing arbitrary initialization data
* `on_destroy`: called when an account is destroyed with **message data** containing arbitrary destruction data
* `on_migrate`: called when an account is migrated to a new code handler. Such handlers receive structured **message data** specifying the old code handler so that the account can perform migration operations or return an error if migration is not possible.

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

The **hypervisor** as a first-class **module** itself contains stateful mappings for:
* **account address** to **handler id** and **account config** 
* **module name** to module **account address**
* **message name** to **account address** for **module messages**

Each **virtual machine** is expected to expose a method which takes a **handler id**
and optional **account config** data and returns a reference to an **account handler**
which can be used to run **messages**.
**Virtual machines** will also receive an `invoke` function
so that their **account handlers** can send messages to other **accounts**.
**Virtual machines** must also implement a method to return the metadata for each **account handler** by **handler id**.

The **hypervisor** module itself handles the following special **module messages** to manage account
creation, destruction, and migration:
* `create(handler id, address?)`: creates a new account in the specified code environment with the specified handler id and optional pre-defined address (if not provided, a new address is generated). The `on_create` message is called if it is implemented by the account.
* `destroy(address)`: deletes the account with the specified address and calls the `on_destroy` message if it is implemented by the account.
* `migrate(address, new handler id, migration data)`: migrates the account with the specified address to the new account handler. The `on_migrate` message must be implemented by the new code and must not return an error for migration to succeed. `migrate` can only be called by the account itself.
* `force_migrate(account address, new handler id, init data, destroy data)`: this can be used when no `on_migrate` handler can perform a proper migration to the new account handler. In this case, `on_destroy` will be called on the old account handler, the account state will be cleared (TODO: how do we safely clear state?), and `on_create` will be called on the new code. This is a destructive operation and should be used with caution.

### Module Lifecycle & Module Messages

For legacy purposes, **modules** have specific lifecycles and **module messages** have special semantics.
A **module handler** cannot be loaded with the `create` message,
but must be loaded by an external call to the **hypervisor**
which includes the **module name** and **module config** bytes (stored as **account config** internally).
The existing `cosmos.app.v1alpha1.Config` can be used for this purpose if desired.

**Module messages** also allow the definition of pre- and post-handlers.
These are special **message handlers** that can only be defined in **module handlers**
and must be prefixed by the `module:pre:` or `module:post:` prefixes
When modules are loaded in the **hypervisor**, a composite **message handler** will be composed using all the defined
pre- and post-handlers for a given **message name** in the loaded module set.
By default, the ordering will be done alphabetically by **module name**.

### Message Data and Packet Specification

To facilitate efficient cross-language and cross-VM message passing, the precise layout of message packets is important
as it reduces the need for serialization and deserialization in the core **hypervisor** and **virtual machine** layers.

We start by defining a **message packet** as a 64kb (65,536 bytes) array which is aligned to a 64kb boundary.
For most **message handlers** this single packet should be sufficient to contain all the arguments and
return values for a **message**.
In cases where the packet size is too small, additional buffers can be referenced from within the **message packet**.

More details on the specific layout of **message packets** will be specified in a future update to this RFC
or a separate RFC.

### Further Specifications

This specification does not cover many important parts of a complete system such as the encoding of message data,
storage, events, transaction execution, or interaction with consensus environments.
It is the intention of this specification that additional specifications regarding those systems will be layered on
top of this specification.
It may become necessary to include more details regarding specific parts of those systems in this specification
at some point, but as a starting point, this specification is intentionally kept minimal.

## Abandoned Ideas (Optional)

## Decision

TODO

## Consequences (optional)

### Backwards Compatibility

It is intended that existing SDK modules built using `cosmossdk.io/core` can be integrated into this system with
no or minimal changes.

### Positive

TODO

### Negative

TODO

### Neutral

TODO

### References

## Discussion
